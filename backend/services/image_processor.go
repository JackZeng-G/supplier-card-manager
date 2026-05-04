package services

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	MaxFileSize int64 // 最大文件大小（字节）
	TargetWidth int   // 目标宽度
	Quality     int   // JPEG质量 (1-100)
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		MaxFileSize: 5 * 1024 * 1024, // 5MB
		TargetWidth: 1200,            // 目标宽度
		Quality:     85,              // JPEG质量
	}
}

// NeedsProcessing 判断是否需要处理（只有大于5MB才触发）
func (p *ImageProcessor) NeedsProcessing(fileSize int64) bool {
	return fileSize > p.MaxFileSize
}

// ProcessImage 处理图片（仅当大于5MB时压缩）
func (p *ImageProcessor) ProcessImage(imageData []byte, filename string) ([]byte, string, error) {
	// 如果图片小于等于5MB，直接返回原数据
	if int64(len(imageData)) <= p.MaxFileSize {
		return imageData, filename, nil
	}

	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, "", fmt.Errorf("解码图片失败: %v", err)
	}

	// 获取原始尺寸
	bounds := img.Bounds()
	origWidth := bounds.Dx()

	// 调整分辨率：先尝试缩小到原来的70%
	var resizedImg image.Image
	for scale := 0.7; scale >= 0.3; scale -= 0.1 {
		newWidth := int(float64(origWidth) * scale)
		resizedImg = imaging.Resize(img, newWidth, 0, imaging.Lanczos)

		// 编码为JPEG
		var result bytes.Buffer
		err = jpeg.Encode(&result, resizedImg, &jpeg.Options{Quality: p.Quality})
		if err != nil {
			continue
		}

		// 如果压缩后小于5MB，返回结果
		if int64(result.Len()) <= p.MaxFileSize {
			ext := ".jpg"
			newFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ext
			return result.Bytes(), newFilename, nil
		}
	}

	// 如果缩小分辨率后仍然太大，尝试降低质量
	for quality := p.Quality - 10; quality >= 30; quality -= 10 {
		var result bytes.Buffer
		// 使用最小的分辨率
		if resizedImg == nil {
			resizedImg = imaging.Resize(img, int(float64(origWidth)*0.3), 0, imaging.Lanczos)
		}
		err = jpeg.Encode(&result, resizedImg, &jpeg.Options{Quality: quality})
		if err != nil {
			continue
		}

		if int64(result.Len()) <= p.MaxFileSize {
			ext := ".jpg"
			newFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ext
			return result.Bytes(), newFilename, nil
		}
	}

	// 返回最终结果（即使仍然大于5MB）
	ext := ".jpg"
	newFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ext
	var result bytes.Buffer
	jpeg.Encode(&result, resizedImg, &jpeg.Options{Quality: 30})

	return result.Bytes(), newFilename, nil
}

// SaveProcessedImage 保存处理后的图片
func (p *ImageProcessor) SaveProcessedImage(imageData []byte, savePath string) error {
	// 确保目录存在
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(savePath, imageData, 0644)
}
