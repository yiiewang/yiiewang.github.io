package initialize

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"go.uber.org/zap"

	"github.com/yiiewang/yiiewang.github.io/example/20260915/user-service/api/global"
)

func Trans(locale string) error {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			// json tag 按 "," 切分取第一段，如 "mobile,omitempty" → "mobile"
			if s := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]; s == "-" {
				return ""
			} else {
				return s
			}
		})

		zhT := zh.New() // 中文翻译器
		enT := en.New() // 英文翻译器
		// 第一个参数是备用语言环境，后边的参数是应该支持的语言环境
		uni := ut.New(enT, zhT, enT)
		if global.Trans, ok = uni.GetTranslator(locale); !ok {
			zap.S().Errorf("uni.GetTranslator(%s)", locale)
		}

		switch locale {
		case "en":
			en_translations.RegisterDefaultTranslations(v, global.Trans)
		case "zh":
			zh_translations.RegisterDefaultTranslations(v, global.Trans)
		default:
			en_translations.RegisterDefaultTranslations(v, global.Trans)
		}
	}
	return nil
}
