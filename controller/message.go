package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hhr12138/chat_room-message/object"
	"github.com/hhr12138/chat_room-message/service"
	"net/http"
)

func GetMessageByGroupId(ctx *gin.Context){
	request := new(service.MessageRequest)
	err := ctx.ShouldBind(request)
	if err != nil{
		ctx.JSON(http.StatusBadGateway,err.Error())
		return
	}
	if len(request.LastMessageId)==0{
		request.LastMessageId = object.MAX_INF
	}
	result, err := service.GetMessageByGroupId(request)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,err.Error())
		return
	}
	ctx.JSON(http.StatusOK,result)
}

func AddMessageToGroup(ctx *gin.Context){
	request := new(service.Message)
	err := ctx.ShouldBind(request)
	if err != nil{
		ctx.JSON(http.StatusBadGateway,err.Error())
		return
	}
	if request.GroupId==0 || request.TimeStamp == 0{
		ctx.JSON(http.StatusBadGateway,errors.New("parameter mismatch"))
		return
	}
	result, err := service.AddMessageToGroup(request)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,err.Error())
		return
	}
	ctx.JSON(http.StatusOK,result)
}
