package eth

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

func decimalsCall(to common.Address) ethereum.CallMsg {
	return ethereum.CallMsg{
		To:   &to,
		Data: []byte{0x31, 0x3c, 0xe5, 0x67},
	}
}

func symbolCall(to common.Address) ethereum.CallMsg {
	return ethereum.CallMsg{
		To:   &to,
		Data: []byte{0x95, 0xd8, 0x9b, 0x41},
	}
}

func balanceCall(to common.Address, a common.Address) ethereum.CallMsg {
	data := make([]byte, 0, 36)

	data = append(data, 0x70, 0xa0, 0x82, 0x31)
	data = append(data, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0)
	data = append(data, a.Bytes()...)

	return ethereum.CallMsg{
		To:   &to,
		Data: data,
	}
}
