package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"simdokpol/internal/dto"
	"strings"
	"time"

	"golang.org/x/mod/semver" // <-- Import library bawaan Go untuk versi
)

// GANTI DENGAN REPO KAMU YANG SEBENARNYA
const repoOwner = "muhammad1505"
const repoName = "simdokpol-release"

type UpdateService interface {
	CheckForUpdates(currentVersion string) (*dto.UpdateCheckResponse, error)
}

type updateService struct{}

func NewUpdateService() UpdateService {
	return &updateService{}
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
	Body    string `json:"body"`
}

func (s *updateService) CheckForUpdates(currentVersion string) (*dto.UpdateCheckResponse, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "simdokpol-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungi server update: %v", err)
	}
	defer resp.Body.Close()

	// --- PENANGANAN ERROR KHUSUS REPO PRIVATE ---
	if resp.StatusCode == http.StatusNotFound {
		return &dto.UpdateCheckResponse{
			HasUpdate:      false,
			CurrentVersion: currentVersion,
			Error:          "Repo tidak ditemukan atau Private. Pastikan repo publik untuk fitur auto-update.",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("gagal membaca respon server: %v", err)
	}

	// --- PERBAIKAN LOGIKA VERSI (SEMVER) ---
	
	// 1. Pastikan formatnya "v1.0.0" (semver butuh 'v' di depan)
	latestVer := release.TagName
	if !strings.HasPrefix(latestVer, "v") {
		latestVer = "v" + latestVer
	}

	currentVer := currentVersion
	if !strings.HasPrefix(currentVer, "v") {
		currentVer = "v" + currentVer
	}

	// 2. Bandingkan menggunakan library semver
	// Compare return:
	// +1 jika latest > current (Ada Update)
	//  0 jika latest == current
	// -1 jika latest < current
	hasUpdate := false
	
	// Abaikan jika versi dev
	if currentVersion != "dev" && semver.IsValid(latestVer) && semver.IsValid(currentVer) {
		if semver.Compare(latestVer, currentVer) > 0 {
			hasUpdate = true
		}
	} else if currentVersion == "dev" {
		// Opsional: Untuk dev, anggap selalu ada update jika tag valid ditemukan
		// hasUpdate = true 
	}

	return &dto.UpdateCheckResponse{
		HasUpdate:      hasUpdate,
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		DownloadURL:    release.HtmlUrl,
		ReleaseNotes:   release.Body,
	}, nil
}