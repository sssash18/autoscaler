package virtual

import "testing"

func TestReadClusterInfo(t *testing.T) {
	clusterInfoPath := "/tmp"
	shootName := "suyash-local"
	readInitClusterInfo(clusterInfoPath, shootName)
}
