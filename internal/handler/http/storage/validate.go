package storage

import (
	"regexp"
	"strings"

	"github.com/xxcheng123/cloudpan189-share/internal/consts"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/context"
	"github.com/xxcheng123/cloudpan189-share/internal/framework/httpcontext"
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/datatypes"
	"github.com/xxcheng123/cloudpan189-share/internal/repository/models"
	cloudbridgeSvi "github.com/xxcheng123/cloudpan189-share/internal/services/cloudbridge"
)

var reSimpleCode = regexp.MustCompile(`[a-zA-Z0-9]{12,14}`)
var reShareLinkForAdd = regexp.MustCompile(`cloud\.189\.cn\/t\/([a-zA-Z0-9]+)`)

func (h *handler) executeOsTypeSubscribe(ctx context.Context, req *addRequest) (datatypes.JSONMap, httpcontext.BusinessError) {
	if req.OsType != models.OsTypeSubscribe {
		return nil, busCodeStorageOsTypeNotMatch
	}
	if req.SubscribeUser == "" {
		return nil, busCodeStorageSubscribeUserEmpty
	}
	if _, err := h.cloudBridgeService.CheckSubscribeUser(ctx, req.SubscribeUser); err != nil {
		return nil, busCodeStorageQuerySubscribeUserError.WithError(err)
	}
	return datatypes.JSONMap{
		consts.FileAdditionKeyUpUserId: req.SubscribeUser,
	}, nil
}

func (h *handler) executeOsTypeSubscribeShare(ctx context.Context, req *addRequest) (datatypes.JSONMap, string, httpcontext.BusinessError) {
	if req.OsType != models.OsTypeSubscribeShareFolder {
		return nil, "", busCodeStorageOsTypeNotMatch
	}
	if req.SubscribeUser == "" || req.ShareCode == "" {
		return nil, "", busCodeStorageSubscribeShareIncomplete
	}
	shareId, isFolder, fileId, err := h.cloudBridgeService.CheckSubscribeShare(ctx, req.SubscribeUser, req.ShareCode)
	if err != nil {
		return nil, "", busCodeStorageQuerySubscribeShareError.WithError(err)
	}
	return datatypes.JSONMap{
		consts.FileAdditionKeyUpUserId: req.SubscribeUser,
		consts.FileAdditionKeyShareId:  shareId,
		consts.FileAdditionKeyIsFolder: isFolder,
	}, fileId, nil
}

func (h *handler) executeOsTypeShare(ctx context.Context, req *addRequest) (datatypes.JSONMap, string, httpcontext.BusinessError) {
	if req.OsType != models.OsTypeShareFolder {
		return nil, "", busCodeStorageOsTypeNotMatch
	}
	if req.ShareCode == "" {
		return nil, "", busCodeStorageShareCodeEmpty
	}

	cleanCode := req.ShareCode
	cleanCode = strings.ReplaceAll(cleanCode, "（", "(")
	cleanCode = strings.ReplaceAll(cleanCode, "）", ")")
	cleanCode = strings.ReplaceAll(cleanCode, "：", ":")
	cleanCode = strings.ReplaceAll(cleanCode, "访问码", "")
	cleanCode = strings.TrimSpace(cleanCode)

	if req.ShareAccessCode == "" {
		if start := strings.Index(cleanCode, "("); start > -1 {
			end := strings.Index(cleanCode, ")")
			if end > start {
				req.ShareAccessCode = strings.TrimSpace(cleanCode[start+1 : end])
			}
		}
		if req.ShareAccessCode == "" {
			parts := strings.Fields(cleanCode)
			if len(parts) > 1 {
				last := parts[len(parts)-1]
				if len(last) == 4 {
					req.ShareAccessCode = last
				}
			}
		}
	}
	req.ShareAccessCode = strings.TrimSpace(req.ShareAccessCode)

	finalCode := ""
	if matches := reShareLinkForAdd.FindStringSubmatch(cleanCode); len(matches) > 1 {
		finalCode = matches[1]
	} else {
		if match := reSimpleCode.FindString(cleanCode); match != "" {
			finalCode = match
		} else {
			parts := strings.Fields(cleanCode)
			if len(parts) > 0 {
				finalCode = strings.Trim(parts[0], "()")
			}
		}
	}
	req.ShareCode = finalCode
	result, err := h.cloudBridgeService.CheckShare(ctx, req.ShareCode, req.ShareAccessCode)
	if err != nil {
		return nil, "", busCodeStorageQuerySubscribeShareError.WithError(err)
	}
	if result.ShareId == 0 && result.FileId == "" {
		return nil, "", busCodeStorageQuerySubscribeShareError
	}

	return datatypes.JSONMap{
		consts.FileAdditionKeyShareId:    result.ShareId,
		consts.FileAdditionKeyIsFolder:   result.IsFolder,
		consts.FileAdditionKeyShareMode:  result.ShareMode,
		consts.FileAdditionKeyAccessCode: req.ShareAccessCode,
	}, result.FileId, nil
}

func (h *handler) executeOsTypePersonal(ctx context.Context, req *addRequest) httpcontext.BusinessError {
	if req.OsType != models.OsTypePersonFolder {
		return busCodeStorageOsTypeNotMatch
	}
	if req.FileId == "" {
		return busCodeStoragePersonParamsIncomplete
	}
	if req.CloudToken == 0 {
		return busCodeStorageCloudTokenEmpty
	}
	token, err := h.cloudTokenService.Query(ctx, req.CloudToken)
	if err != nil {
		return busCodeStorageCloudTokenNotExist.WithError(err)
	}
	if _, err = h.cloudBridgeService.CheckPerson(ctx, cloudbridgeSvi.NewAuthToken(token.AccessToken, token.ExpiresIn), req.FileId); err != nil {
		return busCodeStoragePersonFileQueryError.WithError(err)
	}
	return nil
}

func (h *handler) executeOsTypeFamily(ctx context.Context, req *addRequest) (datatypes.JSONMap, httpcontext.BusinessError) {
	if req.OsType != models.OsTypeFamilyFolder {
		return nil, busCodeStorageOsTypeNotMatch
	}
	if req.FileId == "" || req.FamilyId == "" {
		return nil, busCodeStorageFamilyParamsIncomplete
	}
	if req.CloudToken == 0 {
		return nil, busCodeStorageCloudTokenEmpty
	}
	token, err := h.cloudTokenService.Query(ctx, req.CloudToken)
	if err != nil {
		return nil, busCodeStorageCloudTokenNotExist.WithError(err)
	}
	if err = h.cloudBridgeService.CheckFamily(ctx, cloudbridgeSvi.NewAuthToken(token.AccessToken, token.ExpiresIn), req.FamilyId, req.FileId); err != nil {
		return nil, busCodeStorageFamilyFileQueryError.WithError(err)
	}
	return datatypes.JSONMap{
		consts.FileAdditionKeyFamilyId: req.FamilyId,
	}, nil
}
