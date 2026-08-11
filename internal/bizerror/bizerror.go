package bizerror

import (
	"fmt"

	"github.com/cloudwego/kitex/pkg/kerrors"
)

const InvalidArgumentCode int32 = 40001

func InvalidArgument(field string) error {
	return kerrors.NewBizStatusError(
		InvalidArgumentCode,
		fmt.Sprintf("%s must be greater than zero", field),
	)
}
