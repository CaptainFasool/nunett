package plugins

import (
	"fmt"
)

func pluginsManager(errCh chan error) {
	i := <-errCh
	fmt.Println(i)
}
