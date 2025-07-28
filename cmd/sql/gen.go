package main

import (
	"github.com/byteflowing/go-common/config"
	"github.com/byteflowing/go-common/orm"
	"gorm.io/gen"
)

func main() {
	c := &orm.Config{}
	if err := config.ReadConfig("./config.yaml", c); err != nil {
		panic(err)
	}
	db := orm.New(c)
	g := gen.NewGenerator(gen.Config{
		OutPath:           "../../pkg/dal/query",
		ModelPkgPath:      "../../pkg/dal/model",
		WithUnitTest:      false,
		FieldNullable:     true,
		FieldCoverable:    true,
		FieldSignable:     true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		Mode:              gen.WithQueryInterface,
	})
	g.UseDB(db)
	g.ApplyBasic(
		g.GenerateModelAs("user_basic", "UserBasic"),
		g.GenerateModelAs("user_auth", "UserAuth"),
		g.GenerateModelAs("user_login_log", "UserLoginLog"),
	)
	g.Execute()
}
