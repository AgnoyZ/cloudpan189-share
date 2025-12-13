package mountpoint

import (
	"regexp"
	"strings"

	"github.com/xxcheng123/cloudpan189-share/internal/bootstrap"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"

	cloudtokenSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudtoken"
	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
	"github.com/xxcheng123/cloudpan189-share/internal/types/topic"

	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, req *CreateRequest) (int64, error)
	Query(ctx context.Context, fileId int64) (*models.MountPoint, error)
	QueryByPath(ctx context.Context, fullPath string) (*models.MountPoint, error)
	List(ctx context.Context, req *ListRequest) ([]*models.MountPoint, error)
	Count(ctx context.Context, req *ListRequest) (int64, error)
	Delete(ctx context.Context, fileId int64) error
    BatchDelete(ctx context.Context, ids []int64) error
	EnableAutoRefresh(ctx context.Context, fileId int64, enable bool) error
	GetAutoRefreshList(ctx context.Context, req *GetAutoRefreshListRequest) ([]*models.MountPoint, error)
	UpdateRefreshConfig(ctx context.Context, fileId int64, config RefreshConfig) error
	ModifyToken(ctx context.Context, fid int64, tokenId int64) error
	BatchParseText(ctx context.Context, req *topic.BatchParseTextRequest) ([]*topic.BatchParseItem, error)
}

type service struct {
    svc                bootstrap.ServiceContext
    cloudTokenService  cloudtokenSvi.Service
    cloudBridgeService cloudbridgeSvi.Service
}

func NewService(
    svc bootstrap.ServiceContext,
    cloudTokenService cloudtokenSvi.Service,
    cloudBridgeService cloudbridgeSvi.Service,
) Service {
    return &service{
        svc:                svc,
        cloudTokenService:  cloudTokenService,
        cloudBridgeService: cloudBridgeService,
    }
}


func (s *service) getDB(ctx context.Context) *gorm.DB {
	return s.svc.GetDB(ctx).Model(new(models.MountPoint))
}

var (
	reFolderID  = regexp.MustCompile(`^\d+$`)
	reShareLink = regexp.MustCompile(`cloud\.189\.cn\/t\/([a-zA-Z0-9]+)`)
)

// 实现 BatchParseText
func (s *service) BatchParseText(ctx context.Context, req *topic.BatchParseTextRequest) ([]*topic.BatchParseItem, error) {
    // 1. 获取 Token 信息 (用于 CheckPerson)
    tokenInfo, err := s.cloudTokenService.Query(ctx, req.CloudToken)
    if err != nil {
        return nil, err
    }
    // 构造 AuthToken
    authToken := cloudbridgeSvi.NewAuthToken(tokenInfo.AccessToken, tokenInfo.ExpiresIn)

    var results []*topic.BatchParseItem
    lines := strings.Split(req.Content, "\n")

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        parts := strings.Fields(line)
        resourceStr := parts[0]
        accessCode := ""
        if len(parts) > 1 {
            accessCode = parts[1]
        }

        // --- 情况A: 分享链接 ---
        if matches := reShareLink.FindStringSubmatch(resourceStr); len(matches) > 1 {
            shareCode := matches[1]

            // 调用 cloudbridge 获取分享详情
            info, err := s.cloudBridgeService.GetShareInfo(ctx, shareCode, accessCode)

            name := ""
            if err == nil && info != nil {
                name = info.Name
            } else {
                name = "未知分享_" + shareCode
            }

            results = append(results, &topic.BatchParseItem{
                Name:            name,
                OsType:          models.OsTypeShareFolder,
                ShareCode:       shareCode,
                ShareAccessCode: accessCode,
            })

        // --- 情况B: 纯数字文件夹ID ---
        } else if reFolderID.MatchString(resourceStr) {
            fileId := resourceStr

            // 调用 cloudbridge 获取文件名
            name, err := s.cloudBridgeService.CheckPerson(ctx, authToken, fileId)

            if err != nil || name == "" {
                name = "未知文件夹_" + fileId
            }

            results = append(results, &topic.BatchParseItem{
                Name:   name,
                OsType: models.OsTypePersonFolder,
                FileId: fileId,
            })
        }
    }

    return results, nil
}
