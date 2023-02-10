package service

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/hhr12138/chat_room-message/object"
	"strconv"
)

//返回Count个消息, 最后一个消息是LastMessageId的前一条, 第一个则是他的前Count条

type MessageRequest struct {
	GroupIds       []int64  `from:"group_id" binding:"required"`
	Count         int64    `from:"count" binding:"required"`
	LastMessageId string `from:"last_message_id"` //为空则是最后一条
}

type MessageType int

type Message struct {
	Id int64 `json:"id" from:"id"`
	TimeStamp int64 `json:"time_stamp" from:"time_stamp""`
	GroupId int64 `json:"group_id" from:"group_id" binding:"required"`
	UserId int64 `json:"user_id" from:"user_id" binding:"required"`
	TargetId int64 `json:"target_id" from:"target_id"`
	Type MessageType `json:"type" from:"type" binding:"required"`
	Value string `json:"value" from:"value" binding:"required"`
}

type MessageResp struct {
	GroupId int64 `json:"group_id"`
	Messages []*Message `json:"messages"`
}

const(
	String MessageType = iota
	Image
	Video
	Audio
)

func AddMessageToGroup(request *Message) (bool, error){
	marshal, _ := json.Marshal(request)
	z := redis.Z{
		Score:  float64(request.TimeStamp),
		Member: string(marshal),
	}
	object.RedisClient.ZAdd(strconv.FormatInt(request.GroupId,10),z)
	return true,nil
}

//todo: 添加上安全校验功能, 确保当前用户可以获取这些群组的信息, 然后并行获取消息和进行校验, 之后过滤掉不合理的消息
func GetMessageByGroupId(request *MessageRequest) ([]*MessageResp, error){
	result := make([]*MessageResp,0)
	groupIds := request.GroupIds
	rangeBy := redis.ZRangeBy{
		Min:request.LastMessageId,
		Max:"-inf",
		Offset: 0,
		Count: request.Count,
	}

	pipeline := object.RedisClient.Pipeline()
	defer pipeline.Close()
	for _,groupId := range groupIds{
		pipeline.ZRevRangeByScore(strconv.FormatInt(groupId, 10), rangeBy)
	}
	exec, err := pipeline.Exec()
	if err != nil{
		return result,err
	}

	for idx,cmder := range exec{
		resp := cmder.(*redis.StringSliceCmd)
		messageStrs,_ := resp.Result()
		result = append(result, getMessageResp(groupIds[idx],messageStrs))
	}
	return result,nil
}

func getMessageResp(groupId int64, messageStrs []string) *MessageResp{
	messages := make([]*Message,0)
	for i := len(messageStrs)-1; i >= 0; i--{
		messageStr := messageStrs[i]
		message := &Message{}
		json.Unmarshal([]byte(messageStr),message)
		messages = append(messages, message)
	}
	return &MessageResp{
		GroupId: groupId,
		Messages: messages,
	}
}