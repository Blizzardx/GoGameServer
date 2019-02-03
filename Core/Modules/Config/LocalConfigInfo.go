package Config

import (
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"os"
)

func Load(fileName string, configObject interface{}) {

	fi, err := os.Open(fileName)
	if err != nil {
		panic("file not found,path===" + fileName)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)

	e := yaml.UnmarshalStrict(fd, configObject)
	if e != nil {
		panic(e)
	}
}
