package shared

import (
	"fmt"
	"net/url"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
)

var (
	SaltKey    string
	BaseURL    string
	EnableAuth bool

	SettingAddition = models.SettingAddition{}
)

var (
	ShareCache = cache.New(5*time.Minute, 10*time.Minute)
)

func JoinDownloadURL(fileId int64, values url.Values) string {
	return fmt.Sprintf("%s%s", BaseURL, fmt.Sprintf(consts.DownloadURLFormat, fileId, values.Encode()))
}
