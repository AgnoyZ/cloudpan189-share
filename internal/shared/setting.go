package shared

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
)

var (
	SaltKey string
	BaseURL string

	SettingAddition = models.SettingAddition{}
)

var (
	ShareCache = cache.New(5*time.Minute, 10*time.Minute)
)
