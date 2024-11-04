package adaptors

import "junoplugin/models"

type AdaptorInterface interface {
	DepositOrWithdraw(event interface{}) (string, models.BigInt, models.BigInt)
}
type Adaptor struct {
	AdaptorInterface
}

type Driver struct {
	driver AdaptorInterface
}

func Init(source string) AdaptorInterface {
	if source == "juno" {
		v := &JunoAdaptor{}
		return v
	}
	return Adaptor{}
}
