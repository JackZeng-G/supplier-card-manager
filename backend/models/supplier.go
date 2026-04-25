package models

import (
	"time"
)

type Supplier struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Source        string    `json:"source" gorm:"size:100"`                           // 来源
	CompanyName   string    `json:"company_name" gorm:"size:200;index"`               // 公司中文名称
	CompanyNameEn string    `json:"company_name_en" gorm:"size:200"`                  // 公司英文名称
	Contact       string    `json:"contact" gorm:"size:50;index"`                     // 联系人
	Position      string    `json:"position" gorm:"size:50"`                          // 职位
	Phone         string    `json:"phone" gorm:"size:50;index"`                       // 联系电话
	Wechat        string    `json:"wechat" gorm:"size:50"`                            // 微信
	Email         string    `json:"email" gorm:"size:100;index"`                      // 邮箱
	QQ            string    `json:"qq" gorm:"size:20"`                                // QQ
	Address       string    `json:"address" gorm:"size:500"`                          // 地址
	Website       string    `json:"website" gorm:"size:200"`                          // 网站
	NvoccNo       string    `json:"nvocc_no" gorm:"size:50"`                          // NVOCC编号
	StaffSize     string    `json:"staff_size" gorm:"size:50"`                        // 人员规模
	TransportType string    `json:"transport_type" gorm:"size:200"`                    // 运输方式(多选逗号分隔)
	Routes        string    `json:"routes" gorm:"size:200"`                           // 优势航线
	ShippingLine  string    `json:"shipping_line" gorm:"size:200"`                    // 船司关系
	Products      string    `json:"products" gorm:"size:500"`                         // 特色产品
	Remark        string    `json:"remark" gorm:"size:500"`                           // 备注
	Status        string    `json:"status" gorm:"size:50;index"`                      // 合作状态
	CardImage     string    `json:"card_image" gorm:"size:500"`                       // 名片正面图片路径
	CardImageBack string    `json:"card_image_back" gorm:"size:500"`                  // 名片反面图片路径
	CreatedAt     time.Time `json:"created_at" gorm:"index"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Supplier) TableName() string {
	return "suppliers"
}