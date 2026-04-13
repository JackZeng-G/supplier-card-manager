package services

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"strings"
	"supplier-card-system/config"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ocr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ocr/v20181119"
)

// 预编译正则表达式（避免每次请求重复编译）
var (
	// 手机号
	reMobilePhone    = regexp.MustCompile(`1[3-9]\d{9}`)
	reMobilePhoneAnchored = regexp.MustCompile(`^1[3-9]\d{9}$`)
	// 邮箱
	reEmail = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	// URL
	reHTTPURL = regexp.MustCompile(`https?://[^\s]+`)
	reWWWURL  = regexp.MustCompile(`www\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(?:\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+`)
	reDomain  = regexp.MustCompile(`[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(?:\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62}){1,3}`)
	// 中文姓名
	reChineseName   = regexp.MustCompile(`^[\x{4e00}-\x{9fa5}]{2,4}$`)
	reChineseChars  = regexp.MustCompile(`[\x{4e00}-\x{9fa5}]`)
	reChineseWord   = regexp.MustCompile(`[\x{4e00}-\x{9fa5}]+`)
	reChineseOnly   = regexp.MustCompile(`^[\x{4e00}-\x{9fa5}]+$`)
	// 英文相关
	reEnglishOnly   = regexp.MustCompile(`^[A-Za-z\s]+$`)
	reEnglishNameCN = regexp.MustCompile(`^([A-Za-z]+)\s+([\x{4e00}-\x{9fa5}]{2,4})$`)
	// 数字提取
	reDigits = regexp.MustCompile(`\d`)
	// QQ
	reQQ      = regexp.MustCompile(`QQ[:：]?\s*(\d{5,12}(?:[/\d]+)?)`)
	reQQAlt   = regexp.MustCompile(`(?:QQ|qq)[:：]?\s*(\d{5,12})`)
	reQQBack  = regexp.MustCompile(`QQ[:：\s]*([1-9]\d{4,12})`)
	// 微信
	reWechat      = regexp.MustCompile(`(?:微信|WeChat|WECHAT)[:：]?\s*(\S+)`)
	reWechatBack  = regexp.MustCompile(`(微信|wechat|WeChat)[:：\s]*([a-zA-Z0-9_-]+)`)
	// 航线
	reRoutes        = regexp.MustCompile(`(?:优势)?航线[:：]?\s*(.+)`)
	reRoutesLabel   = regexp.MustCompile(`^(优势)?航线[:：]?\s*`)
	reRoutesAlt     = regexp.MustCompile(`(?:优势)?航线[:：]?\s*(.*)`)
	// 船公司
	reShipping      = regexp.MustCompile(`(?:合作|代理)船公司?[:：]?\s*(.+)`)
	reShippingAlt   = regexp.MustCompile(`船公司?[:：]?\s*(.+)`)
	// 特色产品
	reProducts      = regexp.MustCompile(`(?:特色产品|主营|服务项目)[:：]?\s*(.+)`)
	// 地址
	reAddressPrefix = regexp.MustCompile(`^地\s*址\s*[:：]?\s*`)
	reAddress       = regexp.MustCompile(`(?:地址|Address|ADD)[:：]?\s*(.+)`)
	// NVOCC
	reNvoccFull  = regexp.MustCompile(`NVOCC\s*(?:NO|No)?[:\.：]?\s*([A-Z0-9\-]+)`)
	reNvoccMOC   = regexp.MustCompile(`MOC\-NV\s*(\d+)`)
	reNvoccDigits = regexp.MustCompile(`(\d{12})`)
	reNvoccGeneric = regexp.MustCompile(`(?:NVOCC|MOC)[\s\-]*([A-Z0-9\-]+)`)
)

type OCRResult struct {
	CompanyName   string `json:"company_name"`
	CompanyNameEn string `json:"company_name_en"`
	Contact       string `json:"contact"`
	Position      string `json:"position"`
	Phone         string `json:"phone"`
	Wechat        string `json:"wechat"`
	Email         string `json:"email"`
	QQ            string `json:"qq"`
	Address       string `json:"address"`
	Website       string `json:"website"`
	NvoccNo       string `json:"nvocc_no"`
	Routes        string `json:"routes"`         // 优势航线
	ShippingLine  string `json:"shipping_line"`  // 船司关系/航司关系
	Products      string `json:"products"`       // 特色产品
	UnmatchedText string `json:"unmatched_text"` // 无法匹配的多余信息
	RawText       string `json:"raw_text"`
}

// RecognizeCard 使用腾讯云名片识别API识别名片图片（正面用）
func RecognizeCard(imageData []byte) (*OCRResult, error) {
	// 如果没有配置腾讯云密钥，返回模拟数据
	if config.AppConfig.TencentSecretID == "" || config.AppConfig.TencentSecretKey == "" {
		return &OCRResult{
			RawText: "请配置腾讯云OCR密钥",
		}, nil
	}

	// 创建腾讯云OCR客户端
	credential := common.NewCredential(
		config.AppConfig.TencentSecretID,
		config.AppConfig.TencentSecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ocr.tencentcloudapi.com"

	client, err := ocr.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return nil, fmt.Errorf("创建OCR客户端失败: %v", err)
	}

	// 调用名片识别专用接口
	request := ocr.NewBusinessCardOCRRequest()
	request.ImageBase64 = common.StringPtr(base64.StdEncoding.EncodeToString(imageData))

	response, err := client.BusinessCardOCR(request)
	if err != nil {
		return nil, fmt.Errorf("名片OCR识别失败: %v", err)
	}

	// 解析名片识别结果
	result := &OCRResult{}

	if response.Response != nil {
		// 提取所有识别到的文本
		var rawText string
		var lines []string

		// 处理名片识别返回的结构化数据
		businessCardInfos := response.Response.BusinessCardInfos

		// 调试：打印名片API返回的原始信息
		log.Printf("BusinessCardOCR response: %d items", len(businessCardInfos))
		for _, info := range businessCardInfos {
			name := ""
			value := ""
			if info.Name != nil {
				name = *info.Name
			}
			if info.Value != nil {
				value = *info.Value
			}
			log.Printf("  Name: %s, Value: %s", name, value)
		}

		for _, info := range businessCardInfos {
			name := ""
			value := ""
			if info.Name != nil {
				name = *info.Name
			}
			if info.Value != nil {
				value = *info.Value
			}

			// 记录原始文本
			if value != "" {
				rawText += name + ": " + value + "\n"
				lines = append(lines, value)
			}

			// 根据字段名映射到结果
			switch name {
			case "姓名":
				result.Contact = value
			case "英文姓名":
				if result.Contact == "" {
					result.Contact = value
				}
			case "职位":
				result.Position = value
			case "英文职位":
				if result.Position == "" {
					result.Position = value
				}
			case "公司":
				result.CompanyName = value
			case "英文公司":
				// 排除邮箱地址被误识别为英文公司名
				if !strings.Contains(value, "@") && !strings.Contains(value, "Email") && result.CompanyNameEn == "" {
					result.CompanyNameEn = value
				} else if !strings.Contains(value, "@") && !strings.Contains(value, "Email") {
					result.CompanyNameEn += " " + value
				}
			case "地址":
				result.Address = value
			case "英文地址":
				if result.Address == "" {
					result.Address = value
				}
			case "手机":
				// 验证是否为有效手机号
				if isValidPhone(value) {
					result.Phone = value
				}
			case "电话":
				// 验证是否为有效手机号，避免NVOCC号被误识别
				if result.Phone == "" && isValidPhone(value) {
					result.Phone = value
				} else if !isValidPhone(value) {
					// 可能是NVOCC号或其他编号
					if result.NvoccNo == "" {
						// 检查是否包含NVOCC或MOC关键词
						upperValue := strings.ToUpper(value)
						if strings.Contains(upperValue, "NVOCC") || strings.Contains(upperValue, "MOC") {
							result.NvoccNo = value
						} else {
							// 可能是纯数字的NVOCC号（12位数字）
							digits := reDigits.FindAllString(value, -1)
							if len(digits) >= 10 {
								result.NvoccNo = strings.Join(digits, "")
							}
						}
					}
				}
			case "邮箱":
				result.Email = value
			case "网站":
				result.Website = value
			case "微信":
				result.Wechat = value
			case "QQ":
				result.QQ = value
			case "传真":
				// 传真号码可以忽略或放入备注
			default:
				// 未识别的字段放入未匹配文本
				if value != "" && !isIgnoredText(value) {
					if result.UnmatchedText != "" {
						result.UnmatchedText += "\n"
					}
					result.UnmatchedText += name + ": " + value
				}
			}
		}

		// 如果公司英文名还没提取，尝试从英文公司中提取
		if result.CompanyNameEn == "" && result.CompanyName != "" {
			// 从原始文本中查找英文公司名
			for _, line := range lines {
				if isEnglishCompanyName(line) && !strings.Contains(line, result.CompanyName) {
					result.CompanyNameEn = line
					break
				}
			}
		}

		// 后处理：提取额外的业务信息（航线、产品等）
		result = postProcessCardResult(result, lines, rawText)

		log.Printf("[DEBUG] After postProcessCardResult, Contact: %s", result.Contact)

		// 判断是否需要通用OCR补充信息
		needsGeneralOCR := result.Contact == "" ||
			strings.Contains(result.Contact, "@") ||
			strings.Contains(result.Contact, ".") ||
			len([]rune(result.Contact)) <= 1 ||
			isJobTitle(result.Contact) ||
			result.Phone == "" ||
			result.Email == "" ||
			result.CompanyName == ""

		if needsGeneralOCR {
			// 先尝试从通用OCR获取完整文字列表
			generalResult, err := recognizeGeneralTextOnly(imageData)
			if err == nil && len(generalResult) > 0 {
				log.Printf("[DEBUG] General OCR returned %d lines: %v", len(generalResult), generalResult)
				// 从通用OCR结果中提取各种字段
				result = extractFieldsFromGeneralOCR(generalResult, result)
			}
			// 如果联系人还是没找到或仍是职位名，尝试从已有lines中提取
			if result.Contact == "" || strings.Contains(result.Contact, "@") || strings.Contains(result.Contact, ".") || len([]rune(result.Contact)) <= 1 || isJobTitle(result.Contact) {
				contact := extractContactNameFromLines(lines, result)
				if contact != "" {
					result.Contact = contact
				}
			}
		}

		result.RawText = rawText
	}

	return result, nil
}

// recognizeGeneralTextOnly 仅获取通用OCR的文字列表
func recognizeGeneralTextOnly(imageData []byte) ([]string, error) {
	if config.AppConfig.TencentSecretID == "" || config.AppConfig.TencentSecretKey == "" {
		return nil, fmt.Errorf("未配置腾讯云密钥")
	}

	credential := common.NewCredential(
		config.AppConfig.TencentSecretID,
		config.AppConfig.TencentSecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ocr.tencentcloudapi.com"

	client, err := ocr.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return nil, err
	}

	request := ocr.NewGeneralAccurateOCRRequest()
	request.ImageBase64 = common.StringPtr(base64.StdEncoding.EncodeToString(imageData))

	response, err := client.GeneralAccurateOCR(request)
	if err != nil {
		return nil, err
	}

	var lines []string
	if response.Response != nil {
		for _, textDetection := range response.Response.TextDetections {
			if textDetection.DetectedText != nil {
				lines = append(lines, *textDetection.DetectedText)
			}
		}
	}
	return lines, nil
}

// extractFieldsFromGeneralOCR 从通用OCR结果中提取多个字段
func extractFieldsFromGeneralOCR(lines []string, result *OCRResult) *OCRResult {
	// 提取联系人姓名
	if result.Contact == "" || strings.Contains(result.Contact, "@") || strings.Contains(result.Contact, ".") || len([]rune(result.Contact)) <= 1 {
		contact := extractContactFromGeneralText(lines, result)
		if contact != "" {
			result.Contact = contact
		}
	}

	// 提取手机号（如果缺失）
	if result.Phone == "" {
		phoneRegex := reMobilePhone
		for _, line := range lines {
			if match := phoneRegex.FindString(line); match != "" {
				result.Phone = match
				break
			}
		}
	}

	// 提取邮箱（如果缺失）
	if result.Email == "" {
		emailRegex := reEmail
		for _, line := range lines {
			if match := emailRegex.FindString(line); match != "" {
				result.Email = strings.ToLower(match)
				break
			}
		}
	}

	// 提取公司名（如果缺失）
	if result.CompanyName == "" {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "公司") || strings.Contains(line, "有限") || strings.Contains(line, "代理") {
				if !strings.Contains(line, "地址") && len(line) > 4 && len(line) < 50 {
					result.CompanyName = line
					break
				}
			}
		}
	}

	// 提取地址（如果缺失）
	if result.Address == "" {
		addressKeywords := []string{"地址", "路", "号", "室", "楼", "区", "市", "省"}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			matchCount := 0
			for _, kw := range addressKeywords {
				if strings.Contains(line, kw) {
					matchCount++
				}
			}
			if matchCount >= 2 && len(line) > 5 {
				// 移除"地址:"前缀
				addr := reAddressPrefix.ReplaceAllString(line, "")
				if addr != "" {
					result.Address = addr
					break
				}
			}
		}
	}

	// 提取网站（如果缺失）
	if result.Website == "" {
		websitePatterns := []*regexp.Regexp{
			reHTTPURL,
			reWWWURL,
		}
		for _, pattern := range websitePatterns {
			for _, line := range lines {
				if match := pattern.FindString(line); match != "" {
					result.Website = match
					break
				}
			}
			if result.Website != "" {
				break
			}
		}
	}

	// 提取QQ（如果缺失）
	if result.QQ == "" {
		qqRegex := reQQ
		for _, line := range lines {
			if match := qqRegex.FindStringSubmatch(line); len(match) > 1 {
				result.QQ = match[1]
				break
			}
		}
	}

	// 提取优势航线（如果缺失）
	if result.Routes == "" {
		routeKeywords := []string{
			"港", "航线", "台", "东南亚", "欧洲", "美洲",
			"地中海", "中东", "印巴", "红海", "非洲", "澳洲",
			"北美", "南美", "东亚", "远东", "近洋", "远洋",
			"美加", "美西", "美东", "欧基", "高加索",
		}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || len(line) > 50 {
				continue
			}
			for _, kw := range routeKeywords {
				if strings.Contains(line, kw) {
					// 清理前缀
					cleanLine := reRoutesLabel.ReplaceAllString(line, "")
					if cleanLine != "" && cleanLine != line {
						result.Routes = cleanLine
					} else if cleanLine != "" {
						result.Routes = line
					}
					break
				}
			}
			if result.Routes != "" {
				break
			}
		}
	}

	// 提取船司关系（如果缺失）
	if result.ShippingLine == "" {
		shippingKeywords := []string{
			"COSCO", "中远", "MSK", "马士基", "MSC", "地中海船",
			"CMA", "达飞", "HPL", "赫伯罗特", "ONE", "OOCL", "东方海外",
			"EMC", "长荣", "YML", "阳明", "HMM", "现代", "EVERGREEN",
			"ZIM", "以星", "PIL", "太平", "WHL", "万海", "RCL",
			"SITC", "海丰", "TSL", "德翔",
		}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			lineUpper := strings.ToUpper(line)
			if line == "" || len(line) > 60 {
				continue
			}
			for _, kw := range shippingKeywords {
				if strings.Contains(lineUpper, strings.ToUpper(kw)) || strings.Contains(line, kw) {
					// 清理前缀
					cleanLine := regexp.MustCompile(`^(船公司?|船司|航司|合作船公司?)[:：]?\s*`).ReplaceAllString(line, "")
					if cleanLine != "" {
						result.ShippingLine = cleanLine
					} else {
						result.ShippingLine = line
					}
					break
				}
			}
			if result.ShippingLine != "" {
				break
			}
		}
	}

	// 提取特色产品（如果缺失）
	if result.Products == "" {
		productKeywords := []string{
			"整箱", "拼箱", "FCL", "LCL",
			"空运", "海运", "陆运", "铁运",
			"报关", "报检", "仓储", "拖车",
			"订舱", "保险", "双清", "到门",
			"进口", "出口", "危险品", "冷链",
			"特种箱", "散货", "大件",
		}
		var productParts []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || len(line) > 40 {
				continue
			}
			lineUpper := strings.ToUpper(line)
			for _, kw := range productKeywords {
				if strings.Contains(line, kw) || strings.Contains(lineUpper, kw) {
					productParts = append(productParts, line)
					break
				}
			}
		}
		if len(productParts) > 0 {
			result.Products = strings.Join(productParts, "、")
		}
	}

	return result
}

// extractContactFromGeneralText 从通用OCR结果中提取联系人姓名
func extractContactFromGeneralText(lines []string, result *OCRResult) string {
	// 常见姓氏
	commonSurnames := []string{
		"王", "李", "张", "刘", "陈", "杨", "黄", "赵", "周", "吴",
		"徐", "孙", "马", "胡", "朱", "郭", "何", "罗", "高", "林",
		"顾", "梁", "宋", "郑", "谢", "韩", "唐", "冯", "于", "董",
	}

	// 排除词
	excludeWords := []string{
		"公司", "有限", "代理", "业务", "航线", "产品", "地址", "电话",
		"邮箱", "微信", "海", "空", "运", "物流", "货运", "国际", "中国",
		"服务", "项目", "特色", "优势", "主营", "承接", "代办", "订舱",
		"出口", "进口", "报关", "仓储", "运输", "配送", "专线", "直达",
		"经理", "总监", "主管", "代表", "专员", "助理", "顾问", "销售",
		"手机", "网址", "HTTP", "http", "www", "NVOCC", "NVOC", "MOC",
		"年份", "月份", "年", "月", "日", // 排除时间相关
	}

	// 邮箱前缀处理（用于匹配）
	emailPrefix := ""
	emailParts := []string{}
	if result.Email != "" {
		parts := strings.Split(result.Email, "@")
		if len(parts) == 2 {
			emailPrefix = strings.ToLower(parts[0])
			// 分割邮箱前缀（按 . _ - 分割）
			emailParts = strings.FieldsFunc(emailPrefix, func(r rune) bool {
				return r == '.' || r == '_' || r == '-'
			})
		}
	}

	// 收集所有候选姓名
	var candidates []string

	// 遍历所有行，查找2-3个中文字符的姓名
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 提取所有连续的中文字符
		chineseRegex := reChineseWord
		matches := chineseRegex.FindAllString(line, -1)
		for _, match := range matches {
			match = strings.TrimSpace(match)
			// 长度必须是2-3个字符
			if len([]rune(match)) < 2 || len([]rune(match)) > 3 {
				continue
			}
			// 检查是否包含排除词
			hasExclude := false
			for _, exclude := range excludeWords {
				if match == exclude || strings.Contains(match, exclude) {
					hasExclude = true
					break
				}
			}
			if hasExclude {
				continue
			}
			// 检查是否以常见姓氏开头
			isSurname := false
			for _, surname := range commonSurnames {
				if strings.HasPrefix(match, surname) {
					isSurname = true
					break
				}
			}
			if !isSurname {
				continue
			}
			candidates = append(candidates, match)
		}
	}

	// 如果有邮箱前缀，尝试匹配
	if len(emailParts) > 0 && len(candidates) > 0 {
		for _, candidate := range candidates {
			// 获取每个字符的拼音
			chars := []rune(candidate)
			charPinyins := make([]string, len(chars))
			for i, char := range chars {
				charPinyins[i] = pinyinApprox(string(char))
			}

			// 检查邮箱前缀的各部分是否与姓名拼音匹配
			matchCount := 0
			for _, ep := range emailParts {
				for _, cp := range charPinyins {
					// 完全匹配或包含关系
					if ep == cp || strings.Contains(ep, cp) || strings.Contains(cp, ep) {
						matchCount++
						break
					}
					// 相似度匹配（处理拼音变体）
					if len(ep) >= 2 && len(cp) >= 2 {
						// 检查前两个字符是否相同
						if ep[:min(2, len(ep))] == cp[:min(2, len(cp))] {
							matchCount++
							break
						}
					}
				}
			}

			// 如果至少有一个邮箱部分匹配了姓名拼音
			if matchCount >= 1 {
				return candidate
			}
		}
	}

	// 如果没有邮箱或没有匹配，返回第一个候选
	if len(candidates) > 0 {
		return candidates[0]
	}

	return ""
}

// RecognizeGeneralText 使用腾讯云通用印刷体识别API识别文本（反面用）
func RecognizeGeneralText(imageData []byte) (*OCRResult, error) {
	// 如果没有配置腾讯云密钥，返回模拟数据
	if config.AppConfig.TencentSecretID == "" || config.AppConfig.TencentSecretKey == "" {
		return &OCRResult{
			RawText: "请配置腾讯云OCR密钥",
		}, nil
	}

	// 创建腾讯云OCR客户端
	credential := common.NewCredential(
		config.AppConfig.TencentSecretID,
		config.AppConfig.TencentSecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ocr.tencentcloudapi.com"

	client, err := ocr.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return nil, fmt.Errorf("创建OCR客户端失败: %v", err)
	}

	// 调用通用印刷体识别接口
	request := ocr.NewGeneralAccurateOCRRequest()
	request.ImageBase64 = common.StringPtr(base64.StdEncoding.EncodeToString(imageData))

	response, err := client.GeneralAccurateOCR(request)
	if err != nil {
		return nil, fmt.Errorf("通用OCR识别失败: %v", err)
	}

	// 解析通用识别结果
	result := &OCRResult{}

	if response.Response != nil {
		// 提取所有识别到的文本
		var rawText string
		var lines []string

		for _, textDetection := range response.Response.TextDetections {
			if textDetection.DetectedText != nil {
				text := *textDetection.DetectedText
				rawText += text + "\n"
				lines = append(lines, text)
			}
		}

		result.RawText = rawText
		result.UnmatchedText = rawText

		// 解析名片背面的业务信息
		result = parseBackCardText(result, lines, rawText)
	}

	return result, nil
}

// parseBackCardText 解析名片背面的文本提取结构化信息
func parseBackCardText(result *OCRResult, lines []string, rawText string) *OCRResult {
	fullText := rawText
	matchedLines := make(map[int]bool)

	// 遍历每一行进行分析
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 1. 提取优势航线 (Routes)
		// 检查是否包含航线标题
		if strings.Contains(line, "航线") || strings.Contains(line, "优势") {
			// 尝试多种匹配模式
			routesPatterns := []string{
				`优势航线[:：]?\s*(.*)`,
				`航线[:：]?\s*(.*)`,
				`优势[:：]?\s*(.*)`,
			}
			for _, pattern := range routesPatterns {
				routesRegex := regexp.MustCompile(pattern)
				if match := routesRegex.FindStringSubmatch(line); len(match) > 1 {
					routeContent := strings.TrimSpace(match[1])
					// 如果当前行有内容，直接使用
					if routeContent != "" && routeContent != "航线" && routeContent != "优势航线" && !isSectionTitle(routeContent) {
						result.Routes = cleanRouteContent(routeContent)
						matchedLines[i] = true
						break
					}
					// 如果当前行只有标题，读取下一行或多行
					var routeParts []string
					for j := i + 1; j < len(lines) && j < i+5; j++ {
						nextLine := strings.TrimSpace(lines[j])
						if nextLine == "" {
							continue
						}
						// 遇到其他标题字段则停止
						if isSectionTitleStrict(nextLine) {
							break
						}
						routeParts = append(routeParts, nextLine)
						matchedLines[j] = true
					}
					if len(routeParts) > 0 {
						result.Routes = cleanRouteContent(strings.Join(routeParts, "、"))
						matchedLines[i] = true
					}
					break
				}
			}
		}

		// 2. 提取船司关系/航司关系 (ShippingLine)
		if strings.Contains(line, "船") || strings.Contains(line, "航司") || strings.Contains(line, "船东") {
			shippingPatterns := []string{
				`船公司?[:：]?\s*(.*)`,
				`船司[:：]?\s*(.*)`,
				`航司[:：]?\s*(.*)`,
				`船东[:：]?\s*(.*)`,
				`航线?船东?[:：]?\s*(.*)`,
				`合作船公司?[:：]?\s*(.*)`,
				`代理船公司?[:：]?\s*(.*)`,
			}
			for _, pattern := range shippingPatterns {
				shippingRegex := regexp.MustCompile(pattern)
				if match := shippingRegex.FindStringSubmatch(line); len(match) > 1 && result.ShippingLine == "" {
					shippingContent := strings.TrimSpace(match[1])
					// 如果当前行有内容
					if shippingContent != "" && !isSectionTitle(shippingContent) {
						result.ShippingLine = shippingContent
						matchedLines[i] = true
						break
					}
					// 如果当前行只有标题，读取下一行或多行
					var shippingParts []string
					for j := i + 1; j < len(lines) && j < i+5; j++ {
						nextLine := strings.TrimSpace(lines[j])
						if nextLine == "" {
							continue
						}
						if isSectionTitleStrict(nextLine) {
							break
						}
						shippingParts = append(shippingParts, nextLine)
						matchedLines[j] = true
					}
					if len(shippingParts) > 0 {
						result.ShippingLine = strings.Join(shippingParts, "、")
						matchedLines[i] = true
					}
					break
				}
			}
		}

		// 3. 提取特色产品/主营 (Products)
		if strings.Contains(line, "产品") || strings.Contains(line, "主营") || strings.Contains(line, "特色") {
			productsPatterns := []string{
				`特色产品[:：]?\s*(.*)`,
				`特色[:：]?\s*(.*)`,
				`主营[:：]?\s*(.*)`,
				`优势产品[:：]?\s*(.*)`,
				`服务项目[:：]?\s*(.*)`,
				`经营范围[:：]?\s*(.*)`,
			}
			for _, pattern := range productsPatterns {
				productsRegex := regexp.MustCompile(pattern)
				if match := productsRegex.FindStringSubmatch(line); len(match) > 1 && result.Products == "" {
					productContent := strings.TrimSpace(match[1])
					// 如果当前行有内容
					if productContent != "" && !isSectionTitle(productContent) {
						result.Products = productContent
						matchedLines[i] = true
						break
					}
					// 如果当前行只有标题，读取下一行或多行
					var productParts []string
					for j := i + 1; j < len(lines) && j < i+5; j++ {
						nextLine := strings.TrimSpace(lines[j])
						if nextLine == "" {
							continue
						}
						if isSectionTitleStrict(nextLine) {
							break
						}
						productParts = append(productParts, nextLine)
						matchedLines[j] = true
					}
					if len(productParts) > 0 {
						result.Products = strings.Join(productParts, "、")
						matchedLines[i] = true
					}
					break
				}
			}
		}
	}

	// 第二轮：从全文中智能提取（如果上面的规则没匹配到）
	// 智能提取航线信息（包含港、台、洲等关键词）
	if result.Routes == "" {
		localMatched := make(map[int]bool)
		for k, v := range matchedLines {
			localMatched[k] = v
		}
		result.Routes = extractRoutesFromText(lines, localMatched)
	}

	// 智能提取船公司信息（包含船公司名称）
	if result.ShippingLine == "" {
		localMatched := make(map[int]bool)
		for k, v := range matchedLines {
			localMatched[k] = v
		}
		result.ShippingLine = extractShippingLineFromText(lines, localMatched)
	}

	// 智能提取产品信息
	if result.Products == "" {
		localMatched := make(map[int]bool)
		for k, v := range matchedLines {
			localMatched[k] = v
		}
		result.Products = extractProductsFromText(lines, localMatched)
	}

	// 4. 提取手机号码
	if result.Phone == "" {
		phoneRegex := reMobilePhone
		if match := phoneRegex.FindString(fullText); match != "" && isValidPhone(match) {
			result.Phone = match
		}
	}

	// 5. 提取邮箱
	if result.Email == "" {
		emailRegex := reEmail
		if match := emailRegex.FindString(fullText); match != "" {
			result.Email = match
		}
	}

	// 6. 提取QQ号
	if result.QQ == "" {
		qqRegex := reQQBack
		if match := qqRegex.FindStringSubmatch(fullText); len(match) > 1 {
			result.QQ = match[1]
		}
	}

	// 7. 提取微信号
	if result.Wechat == "" {
		wechatRegex := reWechatBack
		if match := wechatRegex.FindStringSubmatch(fullText); len(match) > 2 {
			result.Wechat = match[2]
		}
	}

	// 8. 提取NVOCC编号
	if result.NvoccNo == "" {
		upperText := strings.ToUpper(fullText)
		if strings.Contains(upperText, "NVOCC") || strings.Contains(upperText, "MOC") {
			nvoccRegex := regexp.MustCompile(`NVOCC\s*(?:NO|No)?[:\.：]?\s*([A-Z0-9\-]+)|MOC\-NV\s*(\d+)`)
			if match := nvoccRegex.FindStringSubmatch(fullText); len(match) > 1 {
				if match[1] != "" {
					result.NvoccNo = match[1]
				} else if match[2] != "" {
					result.NvoccNo = "MOC-NV" + match[2]
				}
			}
		}
	}

	// 收集未匹配的行
	var unmatchedLines []string
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if !matchedLines[i] && line != "" && !isIgnoredLine(line, result) {
			unmatchedLines = append(unmatchedLines, line)
		}
	}
	if len(unmatchedLines) > 0 {
		result.UnmatchedText = strings.Join(unmatchedLines, "\n")
	}

	return result
}

// isEnglishCompanyName 判断是否为英文公司名
func isEnglishCompanyName(text string) bool {
	// 排除邮箱地址
	if strings.Contains(text, "@") {
		return false
	}

	upper := strings.ToUpper(text)
	return (strings.Contains(upper, "CO.,") ||
		strings.Contains(upper, "LTD") ||
		strings.Contains(upper, "INC") ||
		strings.Contains(upper, "CORP")) &&
		len(text) > 10 // 至少10个字符才是有效的公司名
}

// isIgnoredText 判断是否为应该忽略的文本
func isIgnoredText(text string) bool {
	ignorePatterns := []string{"IATA", "WCA", "FIATA", "NVOCC", "SERVICE", "VALUE"}
	upper := strings.ToUpper(text)
	for _, p := range ignorePatterns {
		if strings.Contains(upper, p) && len(text) < 20 {
			return true
		}
	}
	return false
}

// isJobTitle 判断联系人字段是否被误识别为职位名称
func isJobTitle(contact string) bool {
	if contact == "" {
		return false
	}

	// 常见职位关键词
	jobTitles := []string{
		"经理", "总监", "主管", "代表", "专员", "助理", "顾问", "销售",
		"客户经理", "业务经理", "项目经理", "区域经理", "大客户经理", "办经理",
		"业务", "操作", "客服", "财务", "行政", "总裁", "董事", "合伙人",
		"一级代理", "业务代表", "销售代表", "商务代表",
	}

	contactLower := strings.ToLower(contact)
	for _, title := range jobTitles {
		if contact == title || strings.Contains(contactLower, strings.ToLower(title)) {
			return true
		}
	}

	// 如果包含职位关键词但没有中文姓氏，也可能是职位
	hasPositionKeyword := false
	for _, title := range jobTitles {
		if strings.Contains(contact, title) {
			hasPositionKeyword = true
			break
		}
	}

	// 检查是否以常见姓氏开头
	commonSurnames := []string{
		"王", "李", "张", "刘", "陈", "杨", "黄", "赵", "周", "吴",
		"徐", "孙", "马", "胡", "朱", "郭", "何", "罗", "高", "林",
		"顾", "梁", "宋", "郑", "谢", "韩", "唐", "冯", "于", "董",
		"廖", "曾", "彭", "潘", "田", "董", "袁", "蔡", "卢", "沈",
	}

	startsWithSurname := false
	for _, surname := range commonSurnames {
		if strings.HasPrefix(contact, surname) {
			startsWithSurname = true
			break
		}
	}

	// 如果包含职位关键词但不以姓氏开头，很可能是职位名
	if hasPositionKeyword && !startsWithSurname {
		return true
	}

	return false
}

// isValidPhone 验证是否为有效的手机号
func isValidPhone(phone string) bool {
	// 移除所有非数字字符
	digits := reDigits.FindAllString(phone, -1)
	cleanPhone := strings.Join(digits, "")

	// 检查是否为11位手机号（1开头）
	phoneRegex := reMobilePhoneAnchored
	return phoneRegex.MatchString(cleanPhone)
}

// postProcessCardResult 后处理名片识别结果，提取额外业务信息
func postProcessCardResult(result *OCRResult, lines []string, rawText string) *OCRResult {
	// 从未匹配文本和原始文本中提取航线、产品等信息
	fullText := rawText
	if result.UnmatchedText != "" {
		fullText += "\n" + result.UnmatchedText
	}

	// 如果联系人没识别到，从邮箱前缀猜测
	if result.Contact == "" && result.Email != "" {
		parts := strings.Split(result.Email, "@")
		if len(parts) == 2 {
			prefix := parts[0]
			// 尝试从原文中查找可能的中文姓名
			nameRegex := reChineseName
			for _, line := range lines {
				if nameRegex.MatchString(strings.TrimSpace(line)) {
					result.Contact = strings.TrimSpace(line)
					break
				}
			}
			// 如果还是没找到，使用邮箱前缀
			if result.Contact == "" {
				result.Contact = prefix
			}
		}
	}

	// 如果手机号没识别到，从微信中提取（微信和手机号经常相同）
	if result.Phone == "" && result.Wechat != "" {
		if isValidPhone(result.Wechat) {
			result.Phone = result.Wechat
		}
	}

	// 如果手机号还是空的，从全文提取
	if result.Phone == "" {
		phoneRegex := reMobilePhone
		if match := phoneRegex.FindString(fullText); match != "" {
			result.Phone = match
		}
	}

	// 提取NVOCC编号
	if result.NvoccNo == "" {
		// 尝试多种NVOCC格式
		nvoccPatterns := []*regexp.Regexp{
			reNvoccFull,
			reNvoccMOC,
			reNvoccDigits, // 12位数字可能是NVOCC号
		}
		for _, pattern := range nvoccPatterns {
			if match := pattern.FindStringSubmatch(fullText); len(match) > 1 {
				result.NvoccNo = strings.TrimSpace(match[1])
				break
			}
		}
	}

	// 提取优势航线
	if result.Routes == "" {
		// 先尝试正则匹配
		routesRegex := reRoutes
		if match := routesRegex.FindStringSubmatch(fullText); len(match) > 1 {
			routeContent := strings.TrimSpace(match[1])
			if routeContent != "" && len(routeContent) < 100 {
				result.Routes = routeContent
			}
		}
		// 如果正则没匹配，尝试关键词匹配
		if result.Routes == "" {
			routeKeywords := []string{
				"港", "台", "东南亚", "欧洲", "美洲", "地中海", "中东",
				"印巴", "红海", "非洲", "澳洲", "北美", "南美", "美加",
				"美西", "美东", "远东", "近洋", "远洋", "高加索",
			}
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || len(line) > 50 {
					continue
				}
				for _, kw := range routeKeywords {
					if strings.Contains(line, kw) {
						result.Routes = line
						break
					}
				}
				if result.Routes != "" {
					break
				}
			}
		}
	}

	// 提取特色产品
	if result.Products == "" {
		// 先尝试正则匹配
		productsRegex := reProducts
		if match := productsRegex.FindStringSubmatch(fullText); len(match) > 1 {
			productContent := strings.TrimSpace(match[1])
			if productContent != "" && len(productContent) < 100 {
				result.Products = productContent
			}
		}
		// 如果正则没匹配，尝试关键词匹配
		if result.Products == "" {
			productKeywords := []string{
				"整箱", "拼箱", "FCL", "LCL", "空运", "海运", "陆运",
				"报关", "报检", "仓储", "拖车", "订舱", "双清", "到门",
				"进口", "出口", "危险品", "冷链", "特种箱", "散货",
			}
			var productParts []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || len(line) > 40 {
					continue
				}
				lineUpper := strings.ToUpper(line)
				for _, kw := range productKeywords {
					if strings.Contains(line, kw) || strings.Contains(lineUpper, kw) {
						productParts = append(productParts, line)
						break
					}
				}
			}
			if len(productParts) > 0 {
				result.Products = strings.Join(productParts, "、")
			}
		}
	}

	// 提取船司关系/航司关系
	if result.ShippingLine == "" {
		// 先尝试正则匹配
		shippingPatterns := []*regexp.Regexp{
			reShippingAlt,
			regexp.MustCompile(`船司[:：]?\s*(.+)`),
			regexp.MustCompile(`航司[:：]?\s*(.+)`),
			reShipping,
		}
		for _, pattern := range shippingPatterns {
			if match := pattern.FindStringSubmatch(fullText); len(match) > 1 {
				shippingContent := strings.TrimSpace(match[1])
				if shippingContent != "" && len(shippingContent) < 100 {
					result.ShippingLine = shippingContent
					break
				}
			}
		}
		// 如果正则没匹配，尝试关键词匹配
		if result.ShippingLine == "" {
			shippingKeywords := []string{
				"COSCO", "中远", "MSK", "马士基", "MSC", "CMA", "达飞",
				"HPL", "赫伯罗特", "ONE", "OOCL", "东方海外", "EMC", "长荣",
				"YML", "阳明", "HMM", "现代", "EVERGREEN", "ZIM", "以星",
				"PIL", "太平", "WHL", "万海", "SITC", "海丰",
			}
			for _, line := range lines {
				line = strings.TrimSpace(line)
				lineUpper := strings.ToUpper(line)
				if line == "" || len(line) > 60 {
					continue
				}
				for _, kw := range shippingKeywords {
					if strings.Contains(lineUpper, strings.ToUpper(kw)) || strings.Contains(line, kw) {
						result.ShippingLine = line
						break
					}
				}
				if result.ShippingLine != "" {
					break
				}
			}
		}
	}

	// 如果网站还没提取到，从原始文本中提取
	if result.Website == "" {
		// 优先匹配http开头的URL
		httpRegex := reHTTPURL
		if match := httpRegex.FindString(rawText); match != "" {
			result.Website = match
		} else {
			// 匹配www开头的域名
			wwwRegex := reWWWURL
			if match := wwwRegex.FindString(rawText); match != "" {
				result.Website = match
			}
		}
	}

	// 提取英文公司名（如果API没识别到）
	if result.CompanyNameEn == "" {
		for _, line := range lines {
			if isEnglishCompanyName(line) && !strings.Contains(line, result.CompanyName) {
				result.CompanyNameEn = line
				break
			}
		}
	}

	return result
}

// parseCardText 解析名片文本提取结构化信息
func parseCardText(lines []string, rawText string) *OCRResult {
	result := &OCRResult{}
	matchedLines := make(map[int]bool) // 记录已匹配的行

	// 遍历每一行进行分析
	for i, line := range lines {
		line = strings.TrimSpace(line)
		matched := false

		// 提取中文公司名称（包含"公司"关键词）
		if strings.Contains(line, "公司") || strings.Contains(line, "有限") {
			// 排除地址行
			if !strings.Contains(line, "地址") && result.CompanyName == "" {
				result.CompanyName = line
				matched = true
			}
		}

		// 提取英文公司名称（包含CO.,LTD或LTD）
		upperLine := strings.ToUpper(line)
		if (strings.Contains(upperLine, "CO.,LTD") ||
			strings.Contains(upperLine, "CO., LTD") ||
			strings.Contains(upperLine, "LTD") ||
			strings.Contains(upperLine, "INC") ||
			strings.Contains(upperLine, "CORP") ||
			strings.Contains(upperLine, "FREIGHT") ||
			strings.Contains(upperLine, "CARGO") ||
			strings.Contains(upperLine, "LOGISTICS") ||
			strings.Contains(upperLine, "FORWARDING")) &&
			!reChineseChars.MatchString(line) {
			if result.CompanyNameEn == "" {
				// 尝试合并连续的英文行形成完整公司名
				fullName := line
				// 检查前一行是否也是英文公司名的一部分
				if i > 0 {
					prevLine := strings.TrimSpace(lines[i-1])
					if !reChineseChars.MatchString(prevLine) && len(prevLine) > 3 {
						// 如果前一行是英文且不是已匹配的公司名
						if !strings.Contains(result.CompanyName, prevLine) {
							fullName = prevLine + " " + fullName
						}
					}
				}
				result.CompanyNameEn = fullName
				matched = true
			}
		}

		// 提取职位（包含职位关键词）
		positionKeywords := []string{"经理", "销售", "主管", "总监", "代表", "专员", "助理", "顾问", "业务", "操作", "客服", "财务", "行政", "总裁", "董事", "合伙人", "办经理", "一级代理", "代理"}
		// 排除一些误判的关键词组合
		excludePositionPatterns := []string{"国际货运代理", "货运代理", "代理安排", "代理公司"}
		isExcludedPosition := false
		for _, ep := range excludePositionPatterns {
			if strings.Contains(line, ep) {
				isExcludedPosition = true
				break
			}
		}
		if !isExcludedPosition {
			for _, kw := range positionKeywords {
				if strings.Contains(line, kw) {
					if result.Position == "" {
						result.Position = line
						matched = true
					}
					break
				}
			}
		}

		// 提取手机号
		if strings.Contains(line, "手机") || strings.Contains(line, "电话") || strings.Contains(line, "Tel") || strings.Contains(line, "Mobile") {
			digits := reDigits.FindAllString(line, -1)
			phone := strings.Join(digits, "")
			if len(phone) >= 11 {
				re := reMobilePhone
				if match := re.FindString(phone); match != "" {
					result.Phone = match
					matched = true
				}
			}
		}

		// 提取邮箱行
		if strings.Contains(line, "邮箱") || strings.Contains(line, "电邮") || strings.Contains(line, "Email") || strings.Contains(line, "E-mail") || strings.Contains(line, "MAIL") {
			emailRegex := reEmail
			if match := emailRegex.FindString(line); match != "" {
				result.Email = match
				matched = true
			}
		}

		// 提取地址行
		if strings.Contains(line, "地址") || strings.Contains(line, "Address") || strings.Contains(line, "ADD") {
			addrRegex := reAddress
			if match := addrRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Address = match[1]
				matched = true
			} else if i+1 < len(lines) && !strings.Contains(lines[i+1], "手机") {
				result.Address = lines[i+1]
				matched = true
			}
		}

		// 提取微信
		if strings.Contains(line, "微信") || strings.Contains(line, "WeChat") || strings.Contains(line, "WECHAT") {
			wechatRegex := reWechat
			if match := wechatRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Wechat = match[1]
				matched = true
			}
		}

		// 提取QQ
		if strings.Contains(line, "QQ") || strings.Contains(line, "qq") {
			qqRegex := reQQAlt
			if match := qqRegex.FindStringSubmatch(line); len(match) > 1 {
				result.QQ = match[1]
				matched = true
			}
		}

		// 提取优势航线
		if strings.Contains(line, "优势航线") || strings.Contains(line, "航线") {
			routesRegex := reRoutesAlt
			if match := routesRegex.FindStringSubmatch(line); len(match) > 1 {
				routeContent := strings.TrimSpace(match[1])
				// 如果当前行没有内容，读取下一行
				if routeContent == "" && i+1 < len(lines) {
					routeContent = strings.TrimSpace(lines[i+1])
					matchedLines[i+1] = true
				}
				if routeContent != "" {
					result.Routes = routeContent
					matched = true
				}
			} else if i+1 < len(lines) {
				// 没有冒号的情况，直接读取下一行
				result.Routes = strings.TrimSpace(lines[i+1])
				matchedLines[i+1] = true
				matched = true
			}
		}

		// 提取特色产品
		if strings.Contains(line, "特色产品") || strings.Contains(line, "主营") {
			productsRegex := reProducts
			if match := productsRegex.FindStringSubmatch(line); len(match) > 1 {
				productContent := strings.TrimSpace(match[1])
				// 如果当前行没有内容，读取下一行
				if productContent == "" && i+1 < len(lines) {
					productContent = strings.TrimSpace(lines[i+1])
					matchedLines[i+1] = true
				}
				if productContent != "" {
					result.Products = productContent
					matched = true
				}
			} else if i+1 < len(lines) {
				// 没有冒号的情况，直接读取下一行
				result.Products = strings.TrimSpace(lines[i+1])
				matchedLines[i+1] = true
				matched = true
			}
		}

		// 提取NVOCC编号
		if strings.Contains(line, "NVOCC") || strings.Contains(line, "MOC") {
			// 多种NVOCC格式匹配
			nvoccPatterns := []string{
				`NVOCC\s*NO[:\.：]?\s*([A-Z0-9\-\s]+)`,
				`MOC\-NV\s*(\d+)`,
				`(?:NVOCC|MOC)[\s\-]*([A-Z0-9\-]+)`,
			}
			for _, pattern := range nvoccPatterns {
				nvoccRegex := regexp.MustCompile(pattern)
				if match := nvoccRegex.FindStringSubmatch(line); len(match) > 1 {
					result.NvoccNo = strings.TrimSpace(match[1])
					matched = true
					break
				}
			}
		}

		// 标记已匹配的行
		if matched {
			matchedLines[i] = true
		}
	}

	// 如果邮箱还没提取到，从全文提取
	if result.Email == "" {
		emailRegex := reEmail
		if match := emailRegex.FindString(rawText); match != "" {
			result.Email = match
		}
	}

	// 如果手机号还没提取到，从全文提取
	if result.Phone == "" {
		phoneRegex := reMobilePhone
		if match := phoneRegex.FindString(rawText); match != "" {
			result.Phone = match
		}
	}

	// 智能提取联系人姓名（多种策略）
	result.Contact = extractContactName(lines, rawText, result)

	// 提取网站
	if result.Website == "" {
		// 优先匹配http开头的URL
		httpRegex := reHTTPURL
		if match := httpRegex.FindString(rawText); match != "" {
			result.Website = match
		} else {
			// 匹配www开头的域名
			wwwRegex := reWWWURL
			if match := wwwRegex.FindString(rawText); match != "" {
				result.Website = match
			} else {
				// 匹配其他域名，但排除邮箱和常见干扰词
				urlRegex := reDomain
				if matches := urlRegex.FindAllString(rawText, -1); len(matches) > 0 {
					excludeWords := []string{"FORWARDING", "LOGISTICS", "CO.", "LTD", "BRANCH", "INT'L", "INTERNATIONAL", "FREIGHT", "CARGO", "tencent"}
					for _, m := range matches {
						isExcluded := false
						upperM := strings.ToUpper(m)
						for _, ew := range excludeWords {
							if strings.Contains(upperM, ew) {
								isExcluded = true
								break
							}
						}
						// 排除邮箱地址的域名部分
						if !isExcluded && !strings.Contains(rawText, "@"+m) && len(m) > 8 {
							result.Website = m
							break
						}
					}
				}
			}
		}
	}

	// 收集未匹配的行作为多余信息
	var unmatchedLines []string
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if !matchedLines[i] && line != "" {
			// 排除一些常见的干扰信息
			if !isIgnoredLine(line, result) {
				unmatchedLines = append(unmatchedLines, line)
			}
		}
	}
	if len(unmatchedLines) > 0 {
		result.UnmatchedText = strings.Join(unmatchedLines, "\n")
	}

	return result
}

// extractContactName 智能提取联系人姓名
func extractContactName(lines []string, rawText string, result *OCRResult) string {
	nameRegex := reChineseName
	excludeWords := []string{"公司", "有限", "代理", "业务", "航线", "产品", "地址", "电话", "邮箱", "微信", "QQ", "海", "空", "运", "月", "年"}

	// 策略1: 在职位前面找2-4个中文字符（扩展搜索范围到前10行）
	if result.Position != "" {
		for i, line := range lines {
			if strings.TrimSpace(line) == result.Position && i > 0 {
				// 检查前10行
				for j := i - 1; j >= 0 && j >= i-10; j-- {
					prevLine := strings.TrimSpace(lines[j])
					if nameRegex.MatchString(prevLine) {
						// 确保不是其他字段
						isExcluded := false
						for _, ew := range excludeWords {
							if strings.Contains(prevLine, ew) {
								isExcluded = true
								break
							}
						}
						if !isExcluded {
							return prevLine
						}
					}
				}
			}
		}
	}

	// 策略2: 从邮箱前缀猜测姓名
	if result.Email != "" {
		parts := strings.Split(result.Email, "@")
		if len(parts) == 2 {
			prefix := parts[0]
			// 常见格式: firstname.lastname, firstnamelastname
			if strings.Contains(prefix, ".") {
				nameParts := strings.Split(prefix, ".")
				if len(nameParts) >= 1 {
					return guessNameFromEmailPrefix(prefix, lines)
				}
			}
			// 从原文中查找可能的中文姓名
			name := findChineseNameFromPinyin(prefix, lines)
			if name != "" {
				return name
			}
		}
	}

	// 策略3: 查找独立的中文名字行（2-4个中文字符，不包含关键词）
	excludeKeywords := []string{
		"公司", "有限", "经理", "销售", "主管", "地址", "电话", "手机", "邮箱", "微信", "QQ",
		"传真", "网址", "网站", "代理", "业务", "航线", "产品", "优势", "特色", "服务", "合作",
		"一级", "二级", "代办", "订舱", "承接", "出口", "进口", "海运", "空运", "物流",
	}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if nameRegex.MatchString(line) {
			// 检查是否包含排除关键词
			hasKeyword := false
			for _, kw := range excludeKeywords {
				if strings.Contains(line, kw) {
					hasKeyword = true
					break
				}
			}
			if hasKeyword {
				continue
			}

			// 检查前后行上下文
			// 姓名通常在公司名之后、职位之前，或者在手机/邮箱附近
			contextScore := 0

			// 检查前一行
			if i > 0 {
				prevLine := strings.TrimSpace(lines[i-1])
				// 前一行是公司名
				if strings.Contains(prevLine, "公司") || strings.Contains(strings.ToUpper(prevLine), "LTD") {
					contextScore += 2
				}
				// 前一行是英文（可能是英文名）
				if reEnglishOnly.MatchString(prevLine) && len(prevLine) > 3 {
					contextScore += 1
				}
			}

			// 检查后一行
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				// 后一行是职位
				for _, kw := range []string{"经理", "销售", "主管", "总监", "代表", "专员", "助理", "顾问", "业务", "操作"} {
					if strings.Contains(nextLine, kw) {
						contextScore += 2
						break
					}
				}
				// 后一行包含手机或邮箱
				if strings.Contains(nextLine, "手机") || strings.Contains(nextLine, "邮箱") || strings.Contains(nextLine, "Email") {
					contextScore += 1
				}
			}

			// 如果上下文得分足够，返回这个姓名
			if contextScore >= 2 {
				return line
			}
		}
	}

	// 策略4: 查找英文名+中文姓的组合
	englishNameRegex := reEnglishNameCN
	for _, line := range lines {
		if match := englishNameRegex.FindStringSubmatch(strings.TrimSpace(line)); len(match) == 3 {
			return match[2] // 返回中文名部分
		}
	}

	// 策略5: 从手机号行前面查找
	if result.Phone != "" {
		for i, line := range lines {
			if strings.Contains(line, result.Phone) && i > 0 {
				// 检查前几行
				for j := i - 1; j >= 0 && j >= i-3; j-- {
					prevLine := strings.TrimSpace(lines[j])
					if nameRegex.MatchString(prevLine) && prevLine != result.Position {
						// 确保不是排除词
						isExcluded := false
						for _, ew := range excludeKeywords {
							if strings.Contains(prevLine, ew) {
								isExcluded = true
								break
							}
						}
						if !isExcluded {
							return prevLine
						}
					}
				}
			}
		}
	}

	// 策略6: 最后尝试 - 在全文中查找任何2-4个中文字符的行（作为最后手段）
	// 优先查找靠近手机号或邮箱的行
	if result.Phone != "" || result.Email != "" {
		phoneOrEmailLine := -1
		for i, line := range lines {
			if strings.Contains(line, result.Phone) || strings.Contains(line, result.Email) {
				phoneOrEmailLine = i
				break
			}
		}
		if phoneOrEmailLine > 0 {
			// 搜索前后10行范围
			for j := phoneOrEmailLine - 10; j <= phoneOrEmailLine+10; j++ {
				if j >= 0 && j < len(lines) && j != phoneOrEmailLine {
					line := strings.TrimSpace(lines[j])
					if nameRegex.MatchString(line) {
						// 确保不是其他字段
						isExcluded := false
						for _, ew := range excludeWords {
							if strings.Contains(line, ew) {
								isExcluded = true
								break
							}
						}
						if !isExcluded && line != result.Position {
							return line
						}
					}
				}
			}
		}
	}

	return ""
}

// guessNameFromEmailPrefix 从邮箱前缀猜测姓名
func guessNameFromEmailPrefix(prefix string, lines []string) string {
	// 常见的英文名映射
	commonNames := map[string]string{
		"jack":  "杰克",
		"tom":   "汤姆",
		"jerry": "杰瑞",
		"david": "大卫",
		"mike":  "麦克",
		"tony":  "托尼",
		"kevin": "凯文",
		"jason": "杰森",
		"lucy":  "露西",
		"mary":  "玛丽",
		"lily":  "莉莉",
		"amy":   "艾米",
	}

	prefixLower := strings.ToLower(prefix)
	for eng, chn := range commonNames {
		if strings.Contains(prefixLower, eng) {
			// 尝试在原文中找到对应的中文名
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if reChineseName.MatchString(line) {
					return line
				}
			}
			return chn
		}
	}
	return ""
}

// findChineseNameFromPinyin 根据拼音在原文中查找中文名
func findChineseNameFromPinyin(pinyin string, lines []string) string {
	// 简化处理：查找原文中独立的中文姓名
	nameRegex := reChineseName
	for _, line := range lines {
		if nameRegex.MatchString(strings.TrimSpace(line)) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// isIgnoredLine 判断是否为应该忽略的行
func isIgnoredLine(line string, result *OCRResult) bool {
	lineUpper := strings.ToUpper(line)

	// 忽略常见的干扰信息
	ignorePatterns := []string{
		"IATA", "CATA", "FIATA", "LOGISTICS", "INTERNATIONAL",
		"LINKING", "STEPS", "SUPPLY", "CHAIN",
		"YOUR", "PARTNER", "GLOBAL", "WORLDWIDE",
	}

	for _, pattern := range ignorePatterns {
		if lineUpper == pattern {
			return true
		}
	}

	// 忽略已经提取的信息
	if line == result.CompanyName || line == result.CompanyNameEn ||
		line == result.Contact || line == result.Position ||
		line == result.Phone || line == result.Email {
		return true
	}

	// 忽略纯数字或太短的行
	if len(line) < 2 {
		return true
	}

	return false
}

// isSectionTitle 判断是否为节标题（用于背面识别时判断是否继续收集多行内容）
func isSectionTitle(line string) bool {
	// 只有当行是纯标题（很短且包含特定关键词）才认为是标题
	// 避免把包含实际内容的行误判为标题
	line = strings.TrimSpace(line)
	if len(line) > 15 {
		// 超过15个字符，不太可能是纯标题
		return false
	}

	// 必须是纯标题关键词（不包含实际内容）
	pureTitlePatterns := []string{
		"优势航线", "航线", "船公司", "船司", "航司", "船东",
		"特色产品", "特色", "主营", "服务项目", "经营范围",
		"地址", "电话", "手机", "邮箱", "微信", "QQ", "网址",
	}

	for _, pattern := range pureTitlePatterns {
		if line == pattern {
			return true
		}
		// 如果以这些关键词结尾且很短，也可能是标题
		if strings.HasSuffix(line, pattern) && len(line) < len(pattern)+5 {
			return true
		}
	}

	return false
}

// isSectionTitleStrict 更严格的节标题判断（只匹配纯标题行）
func isSectionTitleStrict(line string) bool {
	line = strings.TrimSpace(line)
	// 只匹配非常短的纯标题
	if len(line) > 10 {
		return false
	}

	pureTitles := []string{
		"优势航线", "航线", "船公司", "船司", "航司", "船东",
		"特色产品", "特色", "主营", "服务项目", "经营范围",
		"地址", "电话", "手机", "邮箱", "微信", "QQ", "网址", "NVOCC",
	}

	for _, title := range pureTitles {
		if line == title || strings.HasPrefix(line, title+":") || strings.HasPrefix(line, title+"：") {
			return true
		}
	}

	return false
}

// cleanRouteContent 清理航线内容
func cleanRouteContent(content string) string {
	// 移除常见的干扰词
	cleanPatterns := []string{
		"优势航线", "优势航线:", "优势航线：",
		"航线", "航线:", "航线：",
		"优势", "优势:", "优势：",
	}
	result := content
	for _, pattern := range cleanPatterns {
		result = strings.ReplaceAll(result, pattern, "")
	}
	return strings.TrimSpace(result)
}

// extractRoutesFromText 从文本中智能提取航线信息
func extractRoutesFromText(lines []string, matchedLines map[int]bool) string {
	// 常见的港口、地区关键词
	routeKeywords := []string{
		"港", "港线", "航线", "台", "东南亚", "欧洲", "美洲",
		"地中海", "中东", "印巴", "红海", "非洲", "澳洲",
		"北美", "南美", "东亚", "远东", "近洋", "远洋",
		"美加", "美西", "美东", "欧基", "欧线", "地中海线",
		"中东线", "红海线", "非洲线", "澳洲线",
	}

	var routeParts []string
	for i, line := range lines {
		if matchedLines[i] {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" || len(line) > 30 {
			continue
		}
		// 检查是否包含航线关键词
		for _, kw := range routeKeywords {
			if strings.Contains(line, kw) {
				routeParts = append(routeParts, line)
				matchedLines[i] = true
				break
			}
		}
	}

	if len(routeParts) > 0 {
		return strings.Join(routeParts, "、")
	}
	return ""
}

// extractShippingLineFromText 从文本中智能提取船公司信息
func extractShippingLineFromText(lines []string, matchedLines map[int]bool) string {
	// 常见的船公司名称关键词
	shippingKeywords := []string{
		"COSCO", "中远", "中远海",
		"MSK", "马士基",
		"MSC", "地中海",
		"CMA", "达飞",
		"HPL", "赫伯罗特",
		"ONE", "海洋网联",
		"OOCL", "东方海外",
		"EMC", "长荣",
		"YML", "阳明",
		"HMM", "现代",
		"EVERGREEN", "长荣",
		"ZIM", "以星",
		"PIL", "太平",
		"WHL", "万海",
		"RCL", "宏海",
		"IAL", "运达",
		"SML", "森罗",
		"KMTC", "高丽海运",
		"SITC", "海丰",
		"TSL", "德翔",
		"NOS", "南星",
		"CK", "天敬",
		"COHEUNG", "京汉",
		"DONGYOUNG", "东映",
		"HEUNG-A", "兴亚",
		"JINZHOU", "锦州",
		"PANCON", "泛洲",
		"PAN OCEAN", "泛洋",
		"POS", "浦项",
		"SINOKOR", "长锦",
		"SNKO", "长锦",
	}

	var shippingParts []string
	for i, line := range lines {
		if matchedLines[i] {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" || len(line) > 50 {
			continue
		}
		// 检查是否包含船公司关键词
		for _, kw := range shippingKeywords {
			if strings.Contains(strings.ToUpper(line), kw) {
				shippingParts = append(shippingParts, line)
				matchedLines[i] = true
				break
			}
		}
	}

	if len(shippingParts) > 0 {
		return strings.Join(shippingParts, "、")
	}
	return ""
}

// extractProductsFromText 从文本中智能提取产品信息
func extractProductsFromText(lines []string, matchedLines map[int]bool) string {
	// 常见的产品/服务关键词
	productKeywords := []string{
		"整箱", "拼箱", "FCL", "LCL",
		"空运", "海运", "陆运", "铁运",
		"报关", "报检", "仓储", "拖车",
		"订舱", "保险", "熏蒸", "商检",
		"双清", "到门", "DDU", "DDP",
		"FOB", "CIF", "CFR", "EXW",
		"进口", "出口", "内贸",
		"普货", "危险品", "冷链", "冻柜",
		"特种箱", "开顶", "框架", "平板",
		"散货", "杂货", "大件", "超重",
		"DDP", "DAP", "CPT", "CIP",
	}

	var productParts []string
	for i, line := range lines {
		if matchedLines[i] {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" || len(line) > 40 {
			continue
		}
		// 检查是否包含产品关键词
		for _, kw := range productKeywords {
			if strings.Contains(line, kw) || strings.Contains(strings.ToUpper(line), kw) {
				productParts = append(productParts, line)
				matchedLines[i] = true
				break
			}
		}
	}

	if len(productParts) > 0 {
		return strings.Join(productParts, "、")
	}
	return ""
}

// extractContactNameFromLines 从OCR文本行中提取联系人姓名
// 用于名片API没有识别到姓名时的备用方案
func extractContactNameFromLines(lines []string, result *OCRResult) string {
	// 常见的中文姓氏（用于验证）
	commonSurnames := []string{
		"王", "李", "张", "刘", "陈", "杨", "黄", "赵", "周", "吴",
		"徐", "孙", "马", "胡", "朱", "郭", "何", "罗", "高", "林",
		"顾", "梁", "宋", "郑", "谢", "韩", "唐", "冯", "于", "董",
		"萧", "程", "曹", "袁", "邓", "许", "傅", "沈", "曾", "彭",
		"吕", "苏", "卢", "蒋", "蔡", "贾", "丁", "魏", "薛", "叶",
		"阎", "余", "潘", "杜", "戴", "夏", "钟", "汪", "田", "任",
		"姜", "范", "方", "石", "姚", "谭", "廖", "邹", "熊", "金",
		"陆", "郝", "孔", "白", "崔", "康", "毛", "邱", "秦", "江",
	}

	// 排除词：不是姓名的常见词汇
	excludeWords := []string{
		"公司", "有限", "代理", "业务", "航线", "产品", "地址", "电话",
		"邮箱", "微信", "海", "空", "运", "物流", "货运", "国际", "中国",
		"服务", "项目", "特色", "优势", "主营", "承接", "代办", "订舱",
		"出口", "进口", "报关", "仓储", "运输", "配送", "专线", "直达",
		"全球", "亚洲", "欧洲", "美洲", "澳洲", "非洲",
		"经理", "总监", "主管", "代表", "专员", "助理", "顾问", "销售",
		"部门", "部门经理", "业务员", "操作员",
	}

	// 判断是否可能是姓名
	isPossibleName := func(text string) bool {
		text = strings.TrimSpace(text)
		// 长度检查：姓名通常是2-4个字符
		if len(text) < 2 || len(text) > 4 {
			return false
		}
		// 必须全是中文字符
		chineseRegex := reChineseOnly
		if !chineseRegex.MatchString(text) {
			return false
		}
		// 检查是否包含排除词
		for _, exclude := range excludeWords {
			if strings.Contains(text, exclude) {
				return false
			}
		}
		// 检查是否以常见姓氏开头
		for _, surname := range commonSurnames {
			if strings.HasPrefix(text, surname) {
				return true
			}
		}
		// 如果不是常见姓氏，但是2-3个字符且不包含排除词，也有可能是姓名
		return len(text) <= 3
	}

	// 计算姓名候选的得分（基于上下文）
	calculateScore := func(line string, lineIndex int) int {
		score := 0
		// 检查前一行
		if lineIndex > 0 {
			prevLine := strings.TrimSpace(lines[lineIndex-1])
			// 前一行是公司名
			if strings.Contains(prevLine, "公司") || strings.Contains(prevLine, "有限") {
				score += 3
			}
			// 前一行包含CO/LTD等（英文名前通常是中文名）
			if strings.Contains(strings.ToUpper(prevLine), "CO.") ||
				strings.Contains(strings.ToUpper(prevLine), "LTD") {
				score += 3
			}
		}

		// 检查后一行
		if lineIndex+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[lineIndex+1])
			// 后一行是职位
			positionKeywords := []string{"经理", "总监", "主管", "代表", "专员", "助理", "顾问", "业务", "销售", "操作"}
			for _, kw := range positionKeywords {
				if strings.Contains(nextLine, kw) {
					score += 5 // 职位在姓名后面是最强的信号
					break
				}
			}
			// 后一行是手机/邮箱
			if strings.Contains(nextLine, "手机") || strings.Contains(nextLine, "电话") ||
				strings.Contains(nextLine, "邮箱") || strings.Contains(nextLine, "Email") ||
				strings.Contains(nextLine, "微信") || strings.Contains(nextLine, "WeChat") {
				score += 3
			}
			// 后一行包含手机号格式
			phoneRegex := reMobilePhone
			if phoneRegex.MatchString(nextLine) {
				score += 3
			}
		}

		// 检查该行是否和职位在同一行（姓名 职位）
		positionKeywords := []string{"经理", "总监", "主管", "代表", "专员", "助理", "顾问", "业务", "销售", "操作"}
		for _, kw := range positionKeywords {
			if strings.Contains(line, kw) {
				// 提取职位前的部分作为姓名
				parts := strings.Split(line, kw)
				if len(parts) > 0 {
					namePart := strings.TrimSpace(parts[0])
					if isPossibleName(namePart) {
						return 10 // 同一行姓名+职位是最强的信号
					}
				}
			}
		}

		return score
	}

	// 策略1: 查找同一行包含姓名和职位的情况（如"顾凯经理"）
	positionKeywords := []string{"经理", "总监", "主管", "代表", "专员", "助理", "顾问", "业务", "销售", "操作"}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		for _, kw := range positionKeywords {
			if strings.Contains(line, kw) && !strings.HasPrefix(line, kw) {
				// 提取职位前的部分
				namePart := strings.Split(line, kw)[0]
				namePart = strings.TrimSpace(namePart)
				if isPossibleName(namePart) {
					return namePart
				}
			}
		}
	}

	// 策略2: 在职位行之前查找姓名
	if result.Position != "" {
		for i, line := range lines {
			if strings.TrimSpace(line) == result.Position && i > 0 {
				// 检查前几行
				for j := i - 1; j >= 0 && j >= i-5; j-- {
					prevLine := strings.TrimSpace(lines[j])
					if isPossibleName(prevLine) {
						return prevLine
					}
				}
			}
			// 职位也可能在同一行
			if strings.Contains(line, result.Position) && !strings.HasPrefix(line, result.Position) {
				namePart := strings.Split(line, result.Position)[0]
				namePart = strings.TrimSpace(namePart)
				if isPossibleName(namePart) {
					return namePart
				}
			}
		}
	}

	// 策略3: 查找独立的2-3个中文字符（可能是姓名），根据上下文评分
	type candidate struct {
		name  string
		index int
		score int
	}
	var candidates []candidate

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if isPossibleName(line) {
			score := calculateScore(line, i)
			if score > 0 {
				candidates = append(candidates, candidate{name: line, index: i, score: score})
			}
		}
	}

	// 返回得分最高的候选
	if len(candidates) > 0 {
		bestCandidate := candidates[0]
		for _, c := range candidates {
			if c.score > bestCandidate.score {
				bestCandidate = c
			}
		}
		return bestCandidate.name
	}

	// 策略4: 如果有邮箱，从邮箱前缀反推（结合拼音匹配）
	if result.Email != "" {
		parts := strings.Split(result.Email, "@")
		if len(parts) == 2 {
			prefix := strings.ToLower(parts[0])
			// 尝试匹配拼音
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if isPossibleName(line) && len(line) >= 2 && len(line) <= 3 {
					// 简单的拼音匹配（首字母）
					if len(line) >= 2 {
						// 例如 "顾凯" -> "gukai"
						pinyin := pinyinApprox(line)
						if strings.Contains(prefix, pinyin) || strings.Contains(pinyin, prefix) {
							return line
						}
					}
				}
			}
		}
	}

	// 策略5: 最后尝试 - 在手机号附近查找
	if result.Phone != "" {
		for i, line := range lines {
			if strings.Contains(line, result.Phone) && i > 0 {
				// 检查前几行
				for j := i - 1; j >= 0 && j >= i-5; j-- {
					prevLine := strings.TrimSpace(lines[j])
					if isPossibleName(prevLine) {
						return prevLine
					}
				}
			}
		}
	}

	return ""
}

// pinyinApprox 汉字的近似拼音（用于姓名匹配）
func pinyinApprox(chinese string) string {
	// 简化的拼音映射（只覆盖常见姓氏和名字）
	pinyinMap := map[rune]string{
		'顾': "gu", '凯': "kai",
		'王': "wang", '李': "li", '张': "zhang", '刘': "liu",
		'陈': "chen", '杨': "yang", '黄': "huang", '赵': "zhao",
		'周': "zhou", '吴': "wu", '徐': "xu", '孙': "sun",
		'马': "ma", '胡': "hu", '朱': "zhu", '郭': "guo",
		'何': "he", '罗': "luo", '高': "gao", '林': "lin",
		'梁': "liang", '宋': "song", '郑': "zheng", '谢': "xie",
		'韩': "han", '唐': "tang", '冯': "feng", '于': "yu",
		'董': "dong", '萧': "xiao", '程': "cheng", '曹': "cao",
		'袁': "yuan", '邓': "deng", '许': "xu", '傅': "fu",
		'沈': "shen", '曾': "zeng", '彭': "peng", '吕': "lv",
		'苏': "su", '卢': "lu", '蒋': "jiang", '蔡': "cai",
		'贾': "jia", '丁': "ding", '魏': "wei", '薛': "xue",
		'叶': "ye", '阎': "yan", '余': "yu", '潘': "pan",
		'杜': "du", '戴': "dai", '夏': "xia", '钟': "zhong",
		'汪': "wang", '田': "tian", '任': "ren", '姜': "jiang",
		'范': "fan", '方': "fang", '石': "shi", '姚': "yao",
		'谭': "tan", '廖': "liao", '邹': "zou", '熊': "xiong",
		'金': "jin", '陆': "lu", '郝': "hao", '孔': "kong",
		'白': "bai", '崔': "cui", '康': "kang", '毛': "mao",
		'邱': "qiu", '秦': "qin", '江': "jiang", '史': "shi",
		// 常见名字
		'伟': "wei", '芳': "fang", '娜': "na", '敏': "min",
		'静': "jing", '丽': "li", '强': "qiang", '磊': "lei",
		'军': "jun", '洋': "yang", '勇': "yong", '艳': "yan",
		'杰': "jie", '娟': "juan", '涛': "tao", '明': "ming",
		'超': "chao", '秀': "xiu", '霞': "xia", '平': "ping",
		'刚': "gang", '桂': "gui", '英': "ying", '华': "hua",
		'文': "wen", '梅': "mei", '海': "hai", '波': "bo",
		'健': "jian", '茜': "qian",
		'婷': "ting", '雪': "xue", '慧': "hui", '红': "hong",
	}

	var result string
	for _, r := range chinese {
		if p, ok := pinyinMap[r]; ok {
			result += p
		}
	}
	return result
}
