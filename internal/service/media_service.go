package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"zee-mirror/internal/config"
	"zee-mirror/pkg/i18n"
	"zee-mirror/pkg/utils"
)

type FFProbeOutput struct {
	Streams []FFStream `json:"streams"`
	Format  FFFormat   `json:"format"`
}

type FFStream struct {
	Tags               map[string]string `json:"tags,omitempty"`
	CodecName          string            `json:"codec_name"`
	CodecType          string            `json:"codec_type"`
	DisplayAspectRatio string            `json:"display_aspect_ratio,omitempty"`
	PixFmt             string            `json:"pix_fmt,omitempty"`
	RFrameRate         string            `json:"r_frame_rate,omitempty"`
	BitRate            string            `json:"bit_rate,omitempty"`
	ChannelLayout      string            `json:"channel_layout,omitempty"`
	SampleRate         string            `json:"sample_rate,omitempty"`
	Index              int               `json:"index"`
	Width              int               `json:"width,omitempty"`
	Height             int               `json:"height,omitempty"`
	Channels           int               `json:"channels,omitempty"`
}

type FFFormat struct {
	Filename   string `json:"filename"`
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
	NbStreams  int    `json:"nb_streams"`
}

type MediaService struct {
	Config *config.Config
}

func NewMediaService(cfg *config.Config) *MediaService {
	return &MediaService{
		Config: cfg,
	}
}

func (s *MediaService) HasAudioStream(inputPath string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inputDir := filepath.Dir(inputPath)
	inputName := filepath.Base(inputPath)

	cmd := execCommand(ctx, "ffprobe", "-v", "error", "-select_streams", "a", "-show_entries", "stream=index", "-of", "csv=p=0", inputName)
	cmd.Dir = inputDir
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

func (s *MediaService) GenerateScreenshotsList(inputPath string, count int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	cleanInput := filepath.Clean(inputPath)
	durationCmd := execCommand(ctx, "ffprobe", "-v", "error", "-show_entries", "format=duration", //nolint:gosec
		"-of", "default=noprint_wrappers=1:nokey=1", "file:"+cleanInput)
	durationOutput, err := durationCmd.Output()
	if err != nil {
		return nil, i18n.WrapError(err, "media_err_duration_fetch", err.Error())
	}

	var duration float64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(durationOutput)), "%f", &duration); err != nil {
		return nil, i18n.WrapError(err, "media_err_duration_parse", err.Error())
	}

	if duration <= 0 {
		return nil, i18n.NewError("media_err_duration_invalid")
	}

	interval := duration / float64(count+1)
	baseName := strings.TrimSuffix(cleanInput, filepath.Ext(cleanInput))

	var screenshots []string
	for i := 1; i <= count; i++ {
		timestamp := interval * float64(i)
		outputPath := fmt.Sprintf("%s_ss%d.jpg", baseName, i)
		cleanOutput := filepath.Clean(outputPath)

		shotCtx, shotCancel := context.WithTimeout(context.Background(), 30*time.Second)

		cmd := execCommand(shotCtx, "ffmpeg", "-ss", fmt.Sprintf("%.2f", timestamp), //nolint:gosec
			"-i", "file:"+cleanInput, "-vframes", "1", "-q:v", "2", "file:"+cleanOutput, "-y")

		if err := cmd.Run(); err != nil {
			slog.Warn("Failed to generate screenshot", "index", i, "error", err)
			shotCancel()
			continue
		}
		shotCancel()
		screenshots = append(screenshots, outputPath)
	}

	if len(screenshots) == 0 {
		return nil, i18n.NewError("media_err_no_screenshots")
	}

	return screenshots, nil
}

// FormatMediaInfo renders ffprobe output as a MarkdownV2 report in lang.
func (s *MediaService) FormatMediaInfo(lang, filename string, info FFProbeOutput) string {
	var content strings.Builder
	line := func(key string, args ...any) {
		content.WriteString(i18n.T(lang, key, args...))
		content.WriteByte('\n')
	}

	line("media_info_file", utils.EscapeMarkdownV2Code(filename))
	size, _ := strconv.ParseInt(info.Format.Size, 10, 64)
	line("media_info_size", utils.EscapeMarkdownV2Code(utils.FormatBytes(size)))

	durationSec, _ := strconv.ParseFloat(info.Format.Duration, 64)
	duration := time.Duration(durationSec * float64(time.Second))
	line("media_info_duration", utils.EscapeMarkdownV2Code(utils.FormatDuration(duration)))

	bitrate, _ := strconv.ParseInt(info.Format.BitRate, 10, 64)
	line("media_info_bitrate", utils.EscapeMarkdownV2Code(utils.FormatBytes(bitrate/8)))
	line("media_info_format", utils.EscapeMarkdownV2Code(info.Format.FormatName))

	videoCount := 0
	audioCount := 0
	for _, st := range info.Streams {
		switch st.CodecType {
		case "video":
			videoCount++
			content.WriteByte('\n')
			line("media_info_video_stream", videoCount)
			line("media_info_codec", utils.EscapeMarkdownV2Code(st.CodecName))
			line("media_info_resolution", st.Width, st.Height)
			if st.DisplayAspectRatio != "" {
				line("media_info_aspect", utils.EscapeMarkdownV2Code(st.DisplayAspectRatio))
			}
			line("media_info_pixfmt", utils.EscapeMarkdownV2Code(st.PixFmt))
			fpsParts := strings.Split(st.RFrameRate, "/")
			if len(fpsParts) == 2 {
				f1, _ := strconv.ParseFloat(fpsParts[0], 64)
				f2, _ := strconv.ParseFloat(fpsParts[1], 64)
				if f2 > 0 {
					line("media_info_fps", f1/f2)
				}
			}
		case "audio":
			audioCount++
			content.WriteByte('\n')
			line("media_info_audio_stream", audioCount)
			line("media_info_codec", utils.EscapeMarkdownV2Code(st.CodecName))
			line("media_info_channels", st.Channels, utils.EscapeMarkdownV2(st.ChannelLayout))
			sampleRate, _ := strconv.ParseInt(st.SampleRate, 10, 64)
			line("media_info_sample_rate", float64(sampleRate)/1000.0)
			if st.BitRate != "" {
				abr, _ := strconv.ParseInt(st.BitRate, 10, 64)
				line("media_info_audio_rate", utils.EscapeMarkdownV2Code(utils.FormatBytes(abr/8)))
			}
		}
	}

	content.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━")

	return ProfessionalMessage(i18n.T(lang, "media_info_title"), content.String())
}

func (s *BotService) DownloadTelegramFile(fileID, fileName string) (string, error) {
	file, isOfficial, err := s.GetFileWithFallback(fileID)
	if err != nil {
		return "", err
	}

	if fileName == "" {
		fileName = filepath.Base(file.FilePath)
	}

	if filepath.IsAbs(file.FilePath) {
		translatedPath := strings.Replace(file.FilePath, "/var/lib/telegram-bot-api", s.Config.DownloadDir, 1)
		if _, err := os.Stat(translatedPath); err == nil {
			slog.Info("Using local telegram file", "path", translatedPath)
			return translatedPath, nil
		}
	}

	inputPath := filepath.Join(s.Config.DownloadDir, fileName)
	downloadURL := s.GetFileLink(file, isOfficial)
	slog.Debug("Telegram file download URL", "url", downloadURL)

	if err := s.DownloadFile(downloadURL, inputPath); err != nil {
		return "", err
	}
	return inputPath, nil
}

func (s *BotService) HasAudioStream(inputPath string) (bool, error) {
	return s.Media.HasAudioStream(inputPath)
}

func (s *BotService) GenerateScreenshotsList(inputPath string, count int) ([]string, error) {
	return s.Media.GenerateScreenshotsList(inputPath, count)
}

func (s *BotService) FormatMediaInfo(lang, filename string, info FFProbeOutput) string {
	return s.Media.FormatMediaInfo(lang, filename, info)
}
