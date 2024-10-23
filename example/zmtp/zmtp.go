package main

import (
	"encoding/json"

	"github.com/khulnasoft/khulnasoft-wasm-go-sdk/example/zmtp/pb"
	"github.com/khulnasoft/khulnasoft-wasm-go-sdk/sdk"
	sdkpb "github.com/khulnasoft/khulnasoft-wasm-go-sdk/sdk/pb"
	_ "github.com/wasilibs/nottinygc"
)

//go:generate mkdir -p pb
//go:generate protoc --go_out=./pb --go-vtproto_out=./pb --go-vtproto_opt=features=unmarshal ./demo.proto
func main() {
	sdk.Info("zmtp-plugin loaded")
	parser := ZrpcParser{}
	parser.Parser = interface{}(parser).(sdk.Parser)
	sdk.SetParser(parser)
}

type ZrpcParser struct {
	sdk.DefaultParser
}

func (p ZrpcParser) HookIn() []sdk.HookBitmap {
	return []sdk.HookBitmap{
		sdk.HOOK_POINT_CUSTOM_MESSAGE,
	}
}

func (p ZrpcParser) CustomMessageHookIn() uint64 {
	return sdk.CustomMessageHookProtocol(sdk.PROTOCOL_ZMTP, true)
}

func (p ZrpcParser) onZMTPMessage(payload []byte) sdk.Action {
	var zmtpMsg sdkpb.ZmtpMessage
	if err := zmtpMsg.UnmarshalVT(payload); err != nil {
		return sdk.ActionNext()
	}
	var msgWrapper pb.MessageWrapper
	if err := msgWrapper.UnmarshalVT(zmtpMsg.Payload); err != nil {
		return sdk.ActionNext()
	}
	jsonData, err := json.Marshal(&msgWrapper)
	if err != nil {
		jsonData = nil
	}
	jsonStr := string(jsonData)
	return sdk.ParseActionAbortWithL7Info([]*sdk.L7ProtocolInfo{{
		Resp:  &sdk.Response{},
		Req:   &sdk.Request{},
		Trace: nil,
		Kv: []sdk.KeyVal{
			{
				Key: "json_payload",
				Val: jsonStr,
			},
		},
		L7ProtocolStr: "Protobuf",
	}})
}

func (p ZrpcParser) OnCustomMessage(ctx *sdk.CustomMessageCtx) sdk.Action {
	if ctx.HookPoint == sdk.ProtocolParse && ctx.TypeCode == uint32(sdk.PROTOCOL_ZMTP) {
		return p.onZMTPMessage(ctx.Payload)
	} else {
		return sdk.ActionNext()
	}
}
