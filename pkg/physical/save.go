package physical

import (
	"io/ioutil"
	"path"
	"runtime"
)

func SaveAsPhysicalFile(fileName string, text []byte) {
	filePath := RootDirectory() + "/" + fileName
	err := ioutil.WriteFile(filePath, text, 0644)
	if err != nil {
		panic(err)
	}
}

func RootDirectory() string {
	_, file, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(file))
}
