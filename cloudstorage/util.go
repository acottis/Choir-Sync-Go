package cloudstorage

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Print(string(b))
}
