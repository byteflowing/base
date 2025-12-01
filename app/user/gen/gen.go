package main

import (
	"gorm.io/gen"

	"github.com/byteflowing/base/pkg/config"
	"github.com/byteflowing/base/pkg/db"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
)

func main() {
	c := &configv1.DbConfig{}
	if err := config.ReadProtoConfig("./config.db.yaml", c); err != nil {
		panic(err)
	}
	_db := db.New(c)
	g := gen.NewGenerator(gen.Config{
		OutPath:           "../../dal/query",
		ModelPkgPath:      "../../dal/model",
		WithUnitTest:      false,
		FieldNullable:     true,
		FieldCoverable:    true,
		FieldSignable:     true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		Mode:              gen.WithQueryInterface,
	})
	g.UseDB(_db)
	g.ApplyBasic(
		g.GenerateModelAs("user_basic", "UserBasic"),
		g.GenerateModelAs("user_auth", "UserAuth"),
		g.GenerateModelAs("user_sign_log", "UserSignLog"),
	)
	g.Execute()
}
