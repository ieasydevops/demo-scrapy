package models

type WebPage struct {
	ID   int    `json:"id" db:"id"`
	URL  string `json:"url" db:"url"`
	Name string `json:"name" db:"name"`
}

type Keyword struct {
	ID      int    `json:"id" db:"id"`
	Keyword string `json:"keyword" db:"keyword"`
}

type MonitorConfig struct {
	ID        int    `json:"id" db:"id"`
	WebPageID int    `json:"web_page_id" db:"web_page_id"`
	CrawlTime string `json:"crawl_time" db:"crawl_time"`
	CrawlFreq string `json:"crawl_freq" db:"crawl_freq"`
	Keywords  string `json:"keywords" db:"keywords"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

type SubscribeConfig struct {
	ID        int    `json:"id" db:"id"`
	Email     string `json:"email" db:"email"`
	PushTime  string `json:"push_time" db:"push_time"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

type PushConfig struct {
	ID       int    `json:"id" db:"id"`
	Email    string `json:"email" db:"email"`
	PushTime string `json:"push_time" db:"push_time"`
}

type Announcement struct {
	ID          int    `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	URL         string `json:"url" db:"url"`
	PublishDate string `json:"publish_date" db:"publish_date"`
	Content     string `json:"content" db:"content"`
	CreatedAt   string `json:"created_at" db:"created_at"`
	WebPageID   int    `json:"web_page_id" db:"web_page_id"`
	WebPageName string `json:"web_page_name" db:"web_page_name"`
	Publisher   string `json:"publisher" db:"publisher"`
}
