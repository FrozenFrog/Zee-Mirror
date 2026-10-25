//go:build !linux

package secretfs

func isRAMBacked(string) (bool, error) {
	return false, nil
}
