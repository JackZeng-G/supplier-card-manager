package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"supplier-card-system/config"
	"supplier-card-system/models"
	"supplier-card-system/services"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// allowedMimeTypes 允许的图片MIME类型
var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/jpg":  true,
}

// validateImageMimeType 验证图片的实际MIME类型
func validateImageMimeType(data []byte) (string, bool) {
	// http.DetectContentType 需要至少512字节才能准确检测
	contentType := http.DetectContentType(data)
	return contentType, allowedMimeTypes[contentType]
}

// moveFile 移动文件，支持跨文件系统（先尝试Rename，失败则复制+删除）
func moveFile(src, dst string) error {
	// 先尝试直接重命名（同文件系统时最快）
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Rename失败，使用复制+删除
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		os.Remove(dst) // 清理部分写入的文件
		return err
	}

	// 确保数据写入磁盘
	if err := dstFile.Sync(); err != nil {
		return err
	}

	// 关闭源文件后删除
	srcFile.Close()
	return os.Remove(src)
}

// moveImageToPermanent 将临时图片移动到永久存储目录，并重命名
func moveImageToPermanent(tempFilename string, companyName string, contact string, suffix string) (string, error) {
	if tempFilename == "" {
		return "", nil
	}

	tempPath := filepath.Join(config.AppConfig.TempUploadPath, tempFilename)

	// 检查临时文件是否存在
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		// 文件可能已经在永久目录中，直接返回原文件名
		permPath := filepath.Join(config.AppConfig.ImagePath, tempFilename)
		if _, err := os.Stat(permPath); err == nil {
			return tempFilename, nil
		}
		return "", nil
	}

	// 确保永久目录存在
	if err := os.MkdirAll(config.AppConfig.ImagePath, 0755); err != nil {
		return "", fmt.Errorf("创建图片目录失败: %v", err)
	}

	// 生成新文件名: 公司名称_联系人_正面/反面.扩展名
	ext := filepath.Ext(tempFilename)
	safeCompanyName := sanitizeFilename(companyName)
	safeContact := sanitizeFilename(contact)
	newFilename := fmt.Sprintf("%s_%s_%s%s", safeCompanyName, safeContact, suffix, ext)
	permPath := filepath.Join(config.AppConfig.ImagePath, newFilename)

	// 如果目标文件已存在，添加时间戳
	if _, err := os.Stat(permPath); err == nil {
		newFilename = fmt.Sprintf("%s_%s_%s_%s%s", safeCompanyName, safeContact, suffix, time.Now().Format("150405"), ext)
		permPath = filepath.Join(config.AppConfig.ImagePath, newFilename)
	}

	// 移动并重命名文件
	if err := moveFile(tempPath, permPath); err != nil {
		return "", fmt.Errorf("移动图片失败: %v", err)
	}

	return newFilename, nil
}

// sanitizeFilename 清理文件名中的非法字符
func sanitizeFilename(name string) string {
	// 替换Windows文件名非法字符
	illegal := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range illegal {
		result = strings.ReplaceAll(result, char, "_")
	}
	// 去除首尾空格
	result = strings.TrimSpace(result)
	// 如果为空，使用默认值
	if result == "" {
		result = "未知"
	}
	// 限制长度（按字符数，不是字节数）
	if utf8.RuneCountInString(result) > 50 {
		runes := []rune(result)
		result = string(runes[:50])
	}
	return result
}

// renameImageIfNeeded 当公司名或联系人变更时重命名图片
func renameImageIfNeeded(oldFilename string, newCompany string, newContact string, oldCompany string, oldContact string, suffix string) (string, error) {
	// 检查是否在永久目录中（临时文件跳过）
	if oldFilename == "" {
		return "", nil
	}

	// 如果公司名和联系人名都没变，不需要重命名
	if newCompany == oldCompany && newContact == oldContact {
		return "", nil
	}

	// 生成新的文件名
	ext := filepath.Ext(oldFilename)
	safeNewCompany := sanitizeFilename(newCompany)
	safeNewContact := sanitizeFilename(newContact)
	newFilename := fmt.Sprintf("%s_%s_%s%s", safeNewCompany, safeNewContact, suffix, ext)

	// 如果文件名没变，不需要处理
	if newFilename == oldFilename {
		return "", nil
	}

	// 获取旧文件和新文件的完整路径
	oldPath := filepath.Join(config.AppConfig.ImagePath, oldFilename)
	newPath := filepath.Join(config.AppConfig.ImagePath, newFilename)

	// 检查旧文件是否存在
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return "", nil // 文件不存在，跳过
	}

	// 如果目标文件已存在，添加时间戳
	if _, err := os.Stat(newPath); err == nil {
		newFilename = fmt.Sprintf("%s_%s_%s_%s%s", safeNewCompany, safeNewContact, suffix, time.Now().Format("150405"), ext)
		newPath = filepath.Join(config.AppConfig.ImagePath, newFilename)
	}

	// 重命名文件
	if err := moveFile(oldPath, newPath); err != nil {
		return "", fmt.Errorf("重命名图片失败: %v", err)
	}

	return newFilename, nil
}

// validateSupplier 验证供应商数据
func validateSupplier(supplier models.Supplier) map[string]string {
	errors := make(map[string]string)

	// 验证公司名称或联系人至少有一个（不强制都填）
	if strings.TrimSpace(supplier.CompanyName) == "" && strings.TrimSpace(supplier.Contact) == "" {
		errors["company_name"] = "公司名称和联系人至少填写一个"
	}

	// 验证电话格式（支持手机、座机、国际号码）
	if supplier.Phone != "" {
		// 更宽松的电话格式验证：支持手机、座机、带国际区号的号码
		phoneRegex := regexp.MustCompile(`^[\d\s\-\(\)\+]{7,20}$`)
		if !phoneRegex.MatchString(supplier.Phone) {
			errors["phone"] = "电话格式不正确"
		}
	}

	// 验证邮箱格式
	if supplier.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(supplier.Email) {
			errors["email"] = "邮箱格式不正确"
		}
	}

	// 状态值验证 - 允许中文状态值
	if supplier.Status != "" {
		validStatuses := map[string]bool{
			"active": true, "inactive": true,
			"合作中": true, "待开发": true, "已暂停": true,
		}
		if !validStatuses[supplier.Status] {
			errors["status"] = "状态值无效"
		}
	}

	return errors
}

// UploadCard 上传名片正面并OCR识别
func UploadCard(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传文件"})
		return
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持jpg、jpeg、png格式"})
		return
	}

	// 检查文件大小 (最大10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过10MB"})
		return
	}

	// 读取文件内容
	fileData, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件读取失败"})
		return
	}
	defer fileData.Close()

	imageData, err := io.ReadAll(fileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件读取失败"})
		return
	}

	// 验证实际MIME类型
	mimeType, valid := validateImageMimeType(imageData)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件类型无效，只支持jpg、png图片格式，检测到: " + mimeType})
		return
	}

	// 创建图片处理器
	processor := services.NewImageProcessor()

	// 如果文件较大，进行压缩和裁剪
	var processedData []byte
	var newFilename string
	if processor.NeedsProcessing(file.Size) {
		processedData, newFilename, err = processor.ProcessImage(imageData, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "图片处理失败: " + err.Error()})
			return
		}
	} else {
		processedData = imageData
		newFilename = file.Filename
	}

	// 生成文件名 - 保存到临时目录
	filename := time.Now().Format("20060102150405") + "_" + newFilename
	savePath := filepath.Join(config.AppConfig.TempUploadPath, filename)

	// 保存处理后的文件
	if err := processor.SaveProcessedImage(processedData, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
		return
	}

	// OCR识别（使用处理后的图片）
	ocrResult, err := services.RecognizeCard(processedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OCR识别失败: " + err.Error()})
		return
	}

	// 返回识别结果和文件路径
	c.JSON(http.StatusOK, gin.H{
		"card_image": filename,
		"ocr_result": ocrResult,
	})
}

// UploadCardBack 上传名片反面
func UploadCardBack(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传文件"})
		return
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只支持jpg、jpeg、png格式"})
		return
	}

	// 检查文件大小 (最大10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过10MB"})
		return
	}

	// 读取文件内容
	fileData, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件读取失败"})
		return
	}
	defer fileData.Close()

	imageData, err := io.ReadAll(fileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件读取失败"})
		return
	}

	// 验证实际MIME类型
	mimeType, valid := validateImageMimeType(imageData)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件类型无效，只支持jpg、png图片格式，检测到: " + mimeType})
		return
	}

	// 创建图片处理器
	processor := services.NewImageProcessor()

	// 如果文件较大，进行压缩和裁剪
	var processedData []byte
	var newFilename string
	if processor.NeedsProcessing(file.Size) {
		processedData, newFilename, err = processor.ProcessImage(imageData, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "图片处理失败: " + err.Error()})
			return
		}
	} else {
		processedData = imageData
		newFilename = file.Filename
	}

	// 生成文件名 - 保存到临时目录
	filename := time.Now().Format("20060102150405") + "_back_" + newFilename
	savePath := filepath.Join(config.AppConfig.TempUploadPath, filename)

	// 保存处理后的文件
	if err := processor.SaveProcessedImage(processedData, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
		return
	}

	// OCR识别反面（使用通用文字识别）
	ocrResult, err := services.RecognizeGeneralText(processedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OCR识别失败: " + err.Error()})
		return
	}

	// 返回识别结果和文件路径
	c.JSON(http.StatusOK, gin.H{
		"card_image_back": filename,
		"ocr_result":      ocrResult,
	})
}

// escapeLike 转义LIKE通配符，防止用户输入的%和_被当作通配符
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// GetSuppliers 获取供应商列表
func GetSuppliers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	status := c.Query("status")
	transportType := c.Query("transport_type")

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var suppliers []models.Supplier
	var total int64

	query := models.DB.Model(&models.Supplier{})

	if search != "" {
		escaped := "%" + escapeLike(search) + "%"
		query = query.Where(
			"company_name LIKE ? ESCAPE '\\' OR contact LIKE ? ESCAPE '\\' OR phone LIKE ? ESCAPE '\\' OR email LIKE ? ESCAPE '\\'",
			escaped, escaped, escaped, escaped,
		)
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 运输方式筛选
	if transportType != "" {
		query = query.Where("transport_type = ?", transportType)
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&suppliers)

	c.JSON(http.StatusOK, gin.H{
		"list":      suppliers,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetSupplier 获取单个供应商
func GetSupplier(c *gin.Context) {
	id := c.Param("id")

	var supplier models.Supplier
	if err := models.DB.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "供应商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, supplier)
}

// CreateSupplier 创建供应商
func CreateSupplier(c *gin.Context) {
	var supplier models.Supplier
	if err := c.ShouldBindJSON(&supplier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 验证数据
	if errors := validateSupplier(supplier); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// 移动图片到永久目录
	if supplier.CardImage != "" {
		newPath, err := moveImageToPermanent(supplier.CardImage, supplier.CompanyName, supplier.Contact, "正面")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存名片正面图片失败: " + err.Error()})
			return
		}
		supplier.CardImage = newPath
	}
	if supplier.CardImageBack != "" {
		newPath, err := moveImageToPermanent(supplier.CardImageBack, supplier.CompanyName, supplier.Contact, "反面")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存名片反面图片失败: " + err.Error()})
			return
		}
		supplier.CardImageBack = newPath
	}

	supplier.CreatedAt = time.Now()
	supplier.UpdatedAt = time.Now()

	if err := models.DB.Create(&supplier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusCreated, supplier)
}

// UpdateSupplier 更新供应商
func UpdateSupplier(c *gin.Context) {
	id := c.Param("id")

	var supplier models.Supplier
	if err := models.DB.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "供应商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var updateData models.Supplier
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 验证数据
	if errors := validateSupplier(updateData); len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// 移动新图片到永久目录（临时上传的情况）
	if updateData.CardImage != "" && updateData.CardImage != supplier.CardImage {
		newPath, err := moveImageToPermanent(updateData.CardImage, updateData.CompanyName, updateData.Contact, "正面")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "图片处理失败"})
			return
		}
		updateData.CardImage = newPath
	} else if updateData.CardImage != "" && updateData.CardImage == supplier.CardImage {
		// 图片路径没变，但公司名或联系人可能变了，需要重命名图片
		newPath, err := renameImageIfNeeded(updateData.CardImage, updateData.CompanyName, updateData.Contact, supplier.CompanyName, supplier.Contact, "正面")
		if err == nil && newPath != "" {
			updateData.CardImage = newPath
		}
	}

	if updateData.CardImageBack != "" && updateData.CardImageBack != supplier.CardImageBack {
		newPath, err := moveImageToPermanent(updateData.CardImageBack, updateData.CompanyName, updateData.Contact, "反面")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "图片处理失败"})
			return
		}
		updateData.CardImageBack = newPath
	} else if updateData.CardImageBack != "" && updateData.CardImageBack == supplier.CardImageBack {
		// 图片路径没变，但公司名或联系人可能变了，需要重命名图片
		newPath, err := renameImageIfNeeded(updateData.CardImageBack, updateData.CompanyName, updateData.Contact, supplier.CompanyName, supplier.Contact, "反面")
		if err == nil && newPath != "" {
			updateData.CardImageBack = newPath
		}
	}

	updateData.ID = supplier.ID
	updateData.CreatedAt = supplier.CreatedAt
	updateData.UpdatedAt = time.Now()

	if err := models.DB.Save(&updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, updateData)
}

// DeleteSupplier 删除供应商
func DeleteSupplier(c *gin.Context) {
	id := c.Param("id")

	// 先查询供应商信息，获取图片路径
	var supplier models.Supplier
	if err := models.DB.First(&supplier, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "供应商不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 删除关联的图片文件
	if supplier.CardImage != "" {
		imgPath := filepath.Join(config.AppConfig.ImagePath, supplier.CardImage)
		os.Remove(imgPath)
	}
	if supplier.CardImageBack != "" {
		imgPath := filepath.Join(config.AppConfig.ImagePath, supplier.CardImageBack)
		os.Remove(imgPath)
	}

	if err := models.DB.Delete(&models.Supplier{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetSupplierStats 获取供应商统计数据
func GetSupplierStats(c *gin.Context) {
	var total int64
	var cooperating int64
	var pending int64
	var paused int64

	models.DB.Model(&models.Supplier{}).Count(&total)
	models.DB.Model(&models.Supplier{}).Where("status = ?", "合作中").Count(&cooperating)
	models.DB.Model(&models.Supplier{}).Where("status = ?", "待开发").Count(&pending)
	models.DB.Model(&models.Supplier{}).Where("status = ?", "已暂停").Count(&paused)

	c.JSON(http.StatusOK, gin.H{
		"total":       total,
		"cooperating": cooperating,
		"pending":     pending,
		"paused":      paused,
	})
}

// ExportSuppliers 导出Excel
func ExportSuppliers(c *gin.Context) {
	var suppliers []models.Supplier
	models.DB.Order("created_at DESC").Find(&suppliers)

	// 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 设置工作表名称
	sheetName := "供应商列表"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头
	headers := []string{"序号", "来源", "公司名称", "英文名称", "联系人", "职位",
		"电话", "微信", "邮箱", "QQ", "地址", "网站", "NVOCC编号",
		"人员规模", "运输方式", "优势航线", "船司关系", "特色产品", "备注", "状态", "创建时间"}

	// 写入表头
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetRowStyle(sheetName, 1, 1, headerStyle)

	// 写入数据
	for i, supplier := range suppliers {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), supplier.Source)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), supplier.CompanyName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), supplier.CompanyNameEn)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), supplier.Contact)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), supplier.Position)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), supplier.Phone)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), supplier.Wechat)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), supplier.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), supplier.QQ)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), supplier.Address)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), supplier.Website)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), supplier.NvoccNo)
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), supplier.StaffSize)
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), supplier.TransportType)
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), supplier.Routes)
		f.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), supplier.ShippingLine)
		f.SetCellValue(sheetName, fmt.Sprintf("R%d", row), supplier.Products)
		f.SetCellValue(sheetName, fmt.Sprintf("S%d", row), supplier.Remark)
		f.SetCellValue(sheetName, fmt.Sprintf("T%d", row), supplier.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("U%d", row), supplier.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 6)  // 序号
	f.SetColWidth(sheetName, "B", "B", 10) // 来源
	f.SetColWidth(sheetName, "C", "C", 25) // 公司名称
	f.SetColWidth(sheetName, "D", "D", 25) // 英文名称
	f.SetColWidth(sheetName, "E", "E", 10) // 联系人
	f.SetColWidth(sheetName, "F", "F", 12) // 职位
	f.SetColWidth(sheetName, "G", "G", 15) // 电话
	f.SetColWidth(sheetName, "H", "H", 15) // 微信
	f.SetColWidth(sheetName, "I", "I", 25) // 邮箱
	f.SetColWidth(sheetName, "J", "J", 12) // QQ
	f.SetColWidth(sheetName, "K", "K", 30) // 地址
	f.SetColWidth(sheetName, "L", "L", 20) // 网站
	f.SetColWidth(sheetName, "M", "M", 15) // NVOCC
	f.SetColWidth(sheetName, "N", "N", 10) // 人员规模
	f.SetColWidth(sheetName, "O", "O", 10) // 运输方式
	f.SetColWidth(sheetName, "P", "P", 20) // 优势航线
	f.SetColWidth(sheetName, "Q", "Q", 20) // 船司关系
	f.SetColWidth(sheetName, "R", "R", 25) // 特色产品
	f.SetColWidth(sheetName, "S", "S", 30) // 备注
	f.SetColWidth(sheetName, "T", "T", 10) // 状态
	f.SetColWidth(sheetName, "U", "U", 20) // 创建时间

	// 写入buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成Excel失败"})
		return
	}

	// 设置响应头（RFC 5987编码支持中文文件名）
	filename := fmt.Sprintf("suppliers_%s.xlsx", time.Now().Format("20060102150405"))
	encodedFilename := url.QueryEscape(filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", filename, encodedFilename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
