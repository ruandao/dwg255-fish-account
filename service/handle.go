package service

import (
	"context"
	"github.com/ruandao/dwg255-fish-account/common"
	"github.com/ruandao/dwg255-fish-common/api/thrift/gen-go/rpc"
	"github.com/ruandao/dwg255-fish-common/tools"
	"fmt"
	"github.com/astaxie/beego/logs"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	redisConf = common.AccountConf.RedisConf
	aesTool   *tools.AesEncrypt
)

type UserServer struct {
}

func InitAesTool() {
	var err error
	if aesTool, err = tools.NewAesTool(common.AccountConf.AccountAesKey); err != nil {
		panic("new aes tool err: " + err.Error())
	}
}

func (p *UserServer) GetUserInfoByOpenId(ctx context.Context, openId string) (r *rpc.Result_, err error) {
	logs.Debug("getUserInfoByOpenId openId: %v", openId)
	var existsUserId string
	var userId int
	if existsUserId, err = redisConf.RedisPool.HGet(redisConf.RedisKeyPrefix+"open_id2user_id", openId).Result(); err == nil {
		if userId, err = strconv.Atoi(existsUserId); err == nil {
			return p.GetUserInfoById(ctx, int32(userId))
		}
	} else {
		r = &rpc.Result_{
			Code: rpc.ErrorCode_UserNotExists,
		}
		err = fmt.Errorf("user openId=%v not exists", openId)
	}
	return
}

func (p *UserServer) CreateQQUser(ctx context.Context, userInfo *rpc.UserInfo) (r *rpc.Result_, err error) {
	logs.Debug("CreateQQUser nickName: %v", userInfo.NickName)
	var nextUserId int64
	var existsUserId string
	var userId int
	if existsUserId, err = redisConf.RedisPool.HGet(redisConf.RedisKeyPrefix+"open_id2user_id", userInfo.QqInfo.OpenId).Result(); err == nil {
		if userId, err = strconv.Atoi(existsUserId); err == nil {
			return p.GetUserInfoById(ctx, int32(userId))
		}
	} else {
		nextUserId, err = redisConf.RedisPool.Incr(redisConf.RedisKeyPrefix + "userId").Result()
		token := ""
		registerTime := time.Now()
		if token, err = aesTool.Encrypt(strconv.Itoa(int(nextUserId)) + "-" + strconv.Itoa(int(registerTime.Unix()))); err == nil {
			rand.Seed(time.Now().UnixNano())
			userInfoRedisMap := map[string]interface{}{
				"UserId":        nextUserId,
				"UserName":      userInfo.UserName,
				"NickName":      userInfo.NickName,
				"Sex":           userInfo.Sex,
				"HeadImg":       userInfo.HeadImg,
				"Lv":            userInfo.Lv,
				"Exp":           userInfo.Exp,
				"Vip":           userInfo.Vip, //VIP级别随机给吧,
				"Gems":          userInfo.Gems,
				"RoomId":        0,
				"Power":         userInfo.Power,
				"ReNameCount":   userInfo.ReNameCount,
				"ReHeadCount":   userInfo.ReHeadCount,
				"RegisterDate":  registerTime.Format("2006-01-02 15:04:05"),
				"Ice":           10,
				"Token":         token,
				"openId":        userInfo.QqInfo.OpenId,
				"FigureUrl":     userInfo.QqInfo.FigureUrl,
				"Province":      userInfo.QqInfo.Province,
				"City":          userInfo.QqInfo.City,
				"TotalSpending": userInfo.QqInfo.TotalSpending,
			}
			if _, err = redisConf.RedisPool.HMSet(redisConf.RedisKeyPrefix+strconv.Itoa(int(nextUserId)), userInfoRedisMap).Result(); err == nil {
				if _, err = redisConf.RedisPool.HSet(redisConf.RedisKeyPrefix+"open_id2user_id", userInfo.QqInfo.OpenId, nextUserId).Result(); err == nil {
					userInfo.UserId = nextUserId
					userInfo.Token = token
					r = &rpc.Result_{
						Code:    rpc.ErrorCode_Success,
						UserObj: userInfo,
					}
					return
				}
			}
		}
	}
	return
}

func (p *UserServer) CreateNewUser(ctx context.Context, nickName string, avatarAuto string, gold int64) (r *rpc.Result_, err error) {
	logs.Debug("CreateNewUser nickName: %v", nickName)
	var nextUserId int64
	nextUserId, err = redisConf.RedisPool.Incr(redisConf.RedisKeyPrefix + "userId").Result()
	token := ""
	registerTime := time.Now()
	if token, err = aesTool.Encrypt(strconv.Itoa(int(nextUserId)) + "-" + strconv.Itoa(int(registerTime.Unix()))); err == nil {
		rand.Seed(time.Now().UnixNano())
		vip := int8(rand.Intn(7))
		userInfoRedisMap := map[string]interface{}{
			"UserId":       nextUserId,
			"UserName":     nickName,
			"NickName":     nickName,
			"Sex":          0,
			"HeadImg":      1,
			"Lv":           1,
			"Exp":          0,
			"Vip":          vip, //VIP级别随机给吧,
			"Gems":         gold,
			"RoomId":       0,
			"Power":        0,
			"ReNameCount":  0,
			"ReHeadCount":  0,
			"RegisterDate": registerTime.Format("2006-01-02 15:04:05"),
			"Ice":          10,
			"Token":        token,
		}
		if _, err = redisConf.RedisPool.HMSet(redisConf.RedisKeyPrefix+strconv.Itoa(int(nextUserId)), userInfoRedisMap).Result(); err == nil {

			r = &rpc.Result_{
				Code: rpc.ErrorCode_Success,
				UserObj: &rpc.UserInfo{
					UserId:       nextUserId,
					UserName:     nickName,
					NickName:     nickName,
					Sex:          0,
					HeadImg:      "1",
					Lv:           1,
					Exp:          0,
					Vip:          vip, //VIP级别随机给吧
					Gems:         gold,
					RoomId:       0,
					Power:        0,
					ReNameCount:  0,
					ReHeadCount:  0,
					RegisterDate: registerTime.Format("2006-01-02 15:04:05"),
					Ice:          10,
					Token:        token,
				},
			}
			return
		}
	}
	return
}

func (p *UserServer) GetUserInfoById(ctx context.Context, userId int32) (r *rpc.Result_, err error) {
	logs.Debug("GetUserInfoById: %v", userId)
	result, err := redisConf.RedisPool.HGetAll(redisConf.RedisKeyPrefix + strconv.Itoa(int(userId))).Result()
	if err != nil {
		return r, err
	}
	var Lv, Vip, Gems, RoomId, Power, Ice int
	if Lv, err = strconv.Atoi(result["Lv"]); err != nil {
		return r, err
	}
	if Vip, err = strconv.Atoi(result["Vip"]); err != nil {
		return r, err
	}
	if RoomId, err = strconv.Atoi(result["RoomId"]); err != nil {
		return r, err
	}
	if Gems, err = strconv.Atoi(result["Gems"]); err != nil {
		return r, err
	}
	if Power, err = strconv.Atoi(result["Power"]); err != nil {
		return r, err
	}
	if Ice, err = strconv.Atoi(result["Ice"]); err != nil {
		return r, err
	}

	r = &rpc.Result_{
		Code: rpc.ErrorCode_Success,
		UserObj: &rpc.UserInfo{
			UserId:       int64(userId),
			UserName:     result["UserName"],
			NickName:     result["NickName"],
			Sex:          0,
			HeadImg:      result["HeadImg"],
			Lv:           int32(Lv),
			Exp:          0,
			Vip:          int8(Vip),
			Gems:         int64(Gems),
			RoomId:       int64(RoomId),
			Power:        int64(Power),
			ReNameCount:  0,
			ReHeadCount:  0,
			RegisterDate: result["RegisterDate"],
			Ice:          int64(Ice),
			Token:        result["Token"],
		},
	}
	return
}

func (p *UserServer) GetUserInfoByToken(ctx context.Context, token string) (r *rpc.Result_, err error) {
	logs.Debug("GetUserInfoByToken: %v", token)
	userIdStr := ""
	if userIdStr, err = aesTool.Decrypt(token); err == nil {
		userId := 0
		tokenSplit := strings.Split(userIdStr, "-")
		if len(tokenSplit) > 1 {
			if userId, err = strconv.Atoi(tokenSplit[0]); err == nil {
				return p.GetUserInfoById(context.Background(), int32(userId))
			}
		}
	}
	return
}

func (p *UserServer) ModifyUserInfoById(ctx context.Context, behavior string, userId int32, propType rpc.ModifyPropType, incr int64) (r *rpc.Result_, err error) {
	logs.Debug("ModifyUserInfoById: %v", behavior)
	var exists int64
	userInfoKey := redisConf.RedisKeyPrefix + strconv.Itoa(int(userId))
	if exists, err = common.AccountConf.RedisConf.RedisPool.Exists(userInfoKey).Result(); err == nil {
		if exists == 1 {
			switch propType {
			case rpc.ModifyPropType_gems:
				common.AccountConf.RedisConf.RedisPool.HIncrBy(userInfoKey, "Gems", incr)
			case rpc.ModifyPropType_ice:
				common.AccountConf.RedisConf.RedisPool.HIncrBy(userInfoKey, "Ice", incr)
			case rpc.ModifyPropType_power:
				common.AccountConf.RedisConf.RedisPool.HIncrBy(userInfoKey, "Power", incr)
			case rpc.ModifyPropType_roomId:
				common.AccountConf.RedisConf.RedisPool.HIncrBy(userInfoKey, "RoomId", incr)
			}
			return p.GetUserInfoById(context.Background(), userId)
		}
		err = fmt.Errorf("user [%d] doesnot exists", userId)
	}
	return
}

func (p *UserServer) GetMessage(ctx context.Context, messageType string) (r string, err error) {
	logs.Debug("GetMessage messageType: %v", messageType)
	var redisErr error
	if messageType == "notice" {
		r, redisErr = redisConf.RedisPool.Get(redisConf.RedisKeyPrefix + "notice").Result()
		if r == "" || redisErr != nil {
			r = "个人开发，仅可用于学习研究。"
		}
	} else {
		r, redisErr = redisConf.RedisPool.Get(redisConf.RedisKeyPrefix + "fkgm").Result()
		if r == "" || redisErr != nil {
			r = "爸爸爱你"
		}
	}
	return
}
func (p *UserServer) RenameUserById(ctx context.Context, userId int32, NewName string) (r *rpc.Result_, err error) {
	return
}
