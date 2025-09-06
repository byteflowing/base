package sql_gen

import (
	"gorm.io/gen"

	"github.com/byteflowing/go-common/config"
	"github.com/byteflowing/go-common/orm"
	dbv1 "github.com/byteflowing/proto/gen/go/db/v1"
)

func main() {
	c := &dbv1.DbConfig{}
	if err := config.ReadConfig("./config.db.yaml", c); err != nil {
		panic(err)
	}
	db := orm.New(c)
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
	g.UseDB(db)
	g.ApplyBasic(
		g.GenerateModelAs("user_basic", "UserBasic"),
		g.GenerateModelAs("user_auth", "UserAuth"),
		g.GenerateModelAs("user_sign_log", "UserSignLog"),
	)
	g.Execute()
}
