package captcha

import (
	"github.com/dchest/captcha"
	"github.com/huskar-t/gopher/common/define/cache"
)

func Init(cache cache.Cache) {
	captcha.SetCustomStore(NewStore(cache))
}
