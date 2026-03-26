# TikTok Downloader CLI

Outil en Go qui s'exécute uniquement en terminal pour :
- récupérer les métadonnées d'une vidéo TikTok via l'API TikWM,
- afficher un résumé clair dans la console,
- télécharger la vidéo sans filigrane dans le dossier `downloads/`.

## Prérequis

- Go installé (version 1.25+ recommandée)
- Accès réseau vers `https://www.tikwm.com/api/`

## Installation

Depuis la racine du projet :

```bash
go build -o tiktech .
```

## Utilisation

```bash
./tiktech "https://www.tiktok.com/@username/video/1234567890"
```

## Sortie console

Le CLI affiche :
- la description de la vidéo,
- le titre de la musique,
- les statistiques (vues, likes, partages),
- le lien no-watermark,
- puis un message de succès avec le chemin du fichier téléchargé.

Exemple :

```text
✅ SUCCÈS : Vidéo téléchargée dans -> downloads/1709400000.mp4
```

## Gestion d'erreurs

Le script gère notamment :
- URL manquante ou vide,
- timeout et erreurs réseau API,
- statut HTTP invalide,
- erreur de parsing JSON,
- absence de données métier dans la réponse,
- erreur de téléchargement/écriture fichier.

## Fichiers générés

- `downloads/<timestamp>.mp4`

Le dossier `downloads` est créé automatiquement s'il n'existe pas.