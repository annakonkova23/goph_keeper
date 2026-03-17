package service

import (
	"strconv"

	"github.com/konkovaanna23/gophkeeper/internal/model"

	pb "github.com/konkovaanna23/gophkeeper/pkg/keeperservice"
)

func ConvertTextDataPBToModel(text *pb.TextInfo) *model.TextData {

	return &model.TextData{
		Title: text.GetTitle(),
		Data:  text.GetText(),
		Meta:  text.GetMeta(),
	}
}

func ConvertFileChunkPBToModel(file *pb.FileChunkInfo, number int) *model.FileChunk {

	return &model.FileChunk{
		FileName: file.GetFileName(),
		ChunkNum: number,
		Data:     file.GetData(),
		Meta:     file.GetMeta(),
	}
}

func ConvertBankCardPBToModel(card *pb.BankCardDetails) *model.BankCardData {
	last4 := ""
	if len(card.Number) >= 4 {
		last4 = card.Number[len(card.Number)-4:]
	}
	last4n, _ := strconv.ParseUint(last4, 10, 32)
	return &model.BankCardData{
		Last4:           uint32(last4n),
		NumberEncrypted: card.Number,
		ExpMonth:        card.ExpMonth,
		ExpYear:         card.ExpYear,
		Owner:           card.Owner,
		Meta:            card.Meta,
	}
}

// ConvertTextDataModelToPB converts model.TextData to pb.TextInfo/
func ConvertTextDataModelToPB(text *model.TextData) *pb.TextInfo {
	return &pb.TextInfo{
		Title: text.Title,
		Text:  text.Data,
		Meta:  text.Meta,
	}
}

// ConvertFileChunkModelToPB converts model.FileChunk to pb.FileChunkInfo.
func ConvertFileChunkModelToPB(chunk *model.FileChunk) *pb.FileChunkInfo {
	return &pb.FileChunkInfo{
		FileName: chunk.FileName,
		Data:     chunk.Data,
		Meta:     chunk.Meta,
	}
}

// ConvertBankCardModelToPB converts model.BankCardData to pb.BankCardDetails.
func ConvertBankCardModelToPB(card *model.BankCardData) *pb.BankCardDetails {
	return &pb.BankCardDetails{
		Number:   card.NumberEncrypted,
		ExpMonth: card.ExpMonth,
		ExpYear:  card.ExpYear,
		Owner:    card.Owner,
		Meta:     card.Meta,
	}
}
