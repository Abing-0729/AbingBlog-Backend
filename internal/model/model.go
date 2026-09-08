package model

import (
	"time"

	"gorm.io/gorm"
)

// User 管理员用户（JWT 迭代时启用，先建表备用）。博客只有站长一个用户
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"size:64;uniqueIndex"`
	Password  string `gorm:"size:255"` // bcrypt 哈希，绝不存明文
	CreatedAt time.Time
}

// Category 文章分类
type Category struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:64;uniqueIndex"`
	Slug      string `gorm:"size:64;uniqueIndex"` // URL 友好的英文标识，用于列表筛选
	Sort      int    `gorm:"default:0"`           // 排序权重，越小越靠前
	CreatedAt time.Time
}

// Tag 文章标签（与文章多对多，GORM 自动生成 article_tags 中间表）
type Tag struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:64;uniqueIndex"`
	CreatedAt time.Time
}

// Article 文章：正文存 Markdown 原文，由前端负责渲染
type Article struct {
	ID          uint       `gorm:"primaryKey"`
	Title       string     `gorm:"size:255"`
	Content     string     `gorm:"type:longtext"` // Markdown 原文
	Summary     string     `gorm:"size:500"`
	Cover       string     `gorm:"size:500"`
	CategoryID  uint       // 0 表示未分类
	Category    *Category  // Preload 时填充
	Tags        []Tag      `gorm:"many2many:article_tags;"`
	Status      string     `gorm:"size:16;default:draft"` // draft 草稿 / published 已发布
	ViewCount   int        `gorm:"default:0"`
	PublishedAt *time.Time // 首次发布时间
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"` // 软删除：DELETE 只打标记，数据可恢复
}
