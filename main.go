package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TikWMResponse est la réponse racine de l'API TikWM.
type TikWMResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data *VideoData  `json:"data"`
}

// VideoData contient les métadonnées de la vidéo.
type VideoData struct {
	Title      string      `json:"title"`
	PlayCount  int         `json:"play_count"`
	DiggCount  int         `json:"digg_count"`
	ShareCount int         `json:"share_count"`
	Play       string      `json:"play"`
	MusicInfo  *MusicInfo  `json:"music_info"`
}

// MusicInfo contient les infos de la musique utilisée.
type MusicInfo struct {
	Title string `json:"title"`
}

const (
	baseURL                 = "https://www.tikwm.com/api/"
	usageMsg                = "Usage: tiktech <tiktok_url>"
	httpTimeout             = 30 * time.Second
	downloadTimeout         = 5 * time.Minute
	downloadDir             = "downloads"
	defaultUserAgent        = "TikTech-CLI/1.0"
	colorGreen              = "\033[32m"
	colorCyan               = "\033[36m"
	colorYellow             = "\033[33m"
	colorRed                = "\033[31m"
	colorReset              = "\033[0m"
	actionStatsOnly         = "1"
	actionDownloadOnly      = "2"
	actionStatsAndDownload  = "3"
	actionQuit              = "4"
)

// ensureDownloadsDir crée le dossier downloads s'il n'existe pas (0755).
func ensureDownloadsDir() error {
	return os.MkdirAll(downloadDir, 0755)
}

func main() {
	printBanner()

	tiktokURL, err := getTikTokURL()
	if tiktokURL == "" {
		fmt.Fprintln(os.Stderr, usageMsg)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sErreur URL: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	videoData, err := fetchMetadata(tiktokURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sErreur API: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	action, err := promptAction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sErreur menu: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	switch action {
	case actionStatsOnly:
		printVideoSummary(videoData)
	case actionDownloadOnly:
		filename, downloadErr := downloadVideo(videoData.Play)
		if downloadErr != nil {
			fmt.Fprintf(os.Stderr, "%sErreur téléchargement: %v%s\n", colorRed, downloadErr, colorReset)
			os.Exit(1)
		}
		fmt.Printf("%s✅ SUCCÈS : Vidéo téléchargée dans -> %s/%s%s\n", colorGreen, downloadDir, filename, colorReset)
	case actionStatsAndDownload:
		printVideoSummary(videoData)
		filename, downloadErr := downloadVideo(videoData.Play)
		if downloadErr != nil {
			fmt.Fprintf(os.Stderr, "%sErreur téléchargement: %v%s\n", colorRed, downloadErr, colorReset)
			os.Exit(1)
		}
		fmt.Printf("%s✅ SUCCÈS : Vidéo téléchargée dans -> %s/%s%s\n", colorGreen, downloadDir, filename, colorReset)
	case actionQuit:
		fmt.Printf("%sSortie sans action.%s\n", colorYellow, colorReset)
	default:
		fmt.Fprintf(os.Stderr, "%sChoix invalide.%s\n", colorRed, colorReset)
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Printf("%s========================================%s\n", colorCyan, colorReset)
	fmt.Printf("%s         TikTech Downloader CLI         %s\n", colorCyan, colorReset)
	fmt.Printf("%s========================================%s\n", colorCyan, colorReset)
}

func getTikTokURL() (string, error) {
	if len(os.Args) >= 2 {
		argURL := strings.TrimSpace(os.Args[1])
		if argURL == "" {
			return "", errors.New("URL vide")
		}
		return argURL, nil
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Colle l'URL TikTok: ")
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	inputURL := strings.TrimSpace(line)
	if inputURL == "" {
		return "", errors.New("URL vide")
	}
	return inputURL, nil
}

func fetchMetadata(tiktokURL string) (*VideoData, error) {
	apiURL := baseURL + "?url=" + url.QueryEscape(tiktokURL)
	client := &http.Client{Timeout: httpTimeout}

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("construction requête API: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("appel API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statut API %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiResp TikWMResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("parsing JSON API: %w", err)
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API TikWM: %s", apiResp.Msg)
	}
	if apiResp.Data == nil {
		return nil, errors.New("API sans champ data")
	}
	if strings.TrimSpace(apiResp.Data.Play) == "" {
		return nil, errors.New("URL de téléchargement manquante (data.play)")
	}

	return apiResp.Data, nil
}

func promptAction() (string, error) {
	fmt.Printf("\n%sMENU%s\n", colorYellow, colorReset)
	fmt.Println("1) Voir les stats")
	fmt.Println("2) Télécharger la vidéo")
	fmt.Println("3) Stats + téléchargement")
	fmt.Println("4) Quitter")
	fmt.Print("Choix: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func printVideoSummary(d *VideoData) {
	musicTitle := ""
	if d.MusicInfo != nil {
		musicTitle = d.MusicInfo.Title
	}

	fmt.Printf("\n%s📊 STATS DE LA VIDÉO%s\n", colorCyan, colorReset)
	fmt.Printf("📝 Description : %s\n", d.Title)
	fmt.Printf("🎵 Musique : %s\n", musicTitle)
	fmt.Printf("👀 Vues : %d | ❤️ Likes : %d | 🔄 Partages : %d\n", d.PlayCount, d.DiggCount, d.ShareCount)
	fmt.Printf("🔗 Lien No-Watermark : %s\n", d.Play)
}

func downloadVideo(videoURL string) (string, error) {
	if err := ensureDownloadsDir(); err != nil {
		return "", fmt.Errorf("création dossier downloads: %w", err)
	}

	videoReq, err := http.NewRequest(http.MethodGet, videoURL, nil)
	if err != nil {
		return "", fmt.Errorf("construction requête téléchargement: %w", err)
	}
	videoReq.Header.Set("User-Agent", defaultUserAgent)

	videoClient := &http.Client{Timeout: downloadTimeout}
	videoResp, err := videoClient.Do(videoReq)
	if err != nil {
		return "", fmt.Errorf("appel téléchargement: %w", err)
	}
	defer videoResp.Body.Close()

	if videoResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("statut téléchargement %d", videoResp.StatusCode)
	}

	filename := fmt.Sprintf("%d.mp4", time.Now().Unix())
	outPath := filepath.Join(downloadDir, filename)

	f, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("création fichier: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, videoResp.Body); err != nil {
		_ = os.Remove(outPath)
		return "", fmt.Errorf("écriture fichier vidéo: %w", err)
	}

	return filename, nil
}
