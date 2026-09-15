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

// SiteMetric 保存站点级运行指标；ID 固定为 1，便于原子递增。
type SiteMetric struct {
	ID         uint `gorm:"primaryKey"`
	StartCount uint `gorm:"not null;default:0"`
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

// Project 作品/项目：门户「SELECTED WORK」列表的数据源。
// 结构对齐前端 ProjectSummary（slug/name/detail/stack/githubUrl/demoUrl），
// 复用 article 的 status（draft/published）与 sort 排序约定。
type Project struct {
	ID        uint   `gorm:"primaryKey"`
	Slug      string `gorm:"size:64;uniqueIndex"` // URL 友好标识，前端用作 key
	Name      string `gorm:"size:128"`
	Detail    string `gorm:"size:500"` // 一句话简介
	Stack     string `gorm:"size:255"` // 技术栈展示串，如 "GO · VUE · MYSQL"
	GithubURL string `gorm:"size:500"`
	DemoURL   string `gorm:"size:500"`
	Sort      int    `gorm:"default:0"`             // 排序权重，越小越靠前
	Status    string `gorm:"size:16;default:draft"` // draft 草稿 / published 已发布
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除
}
