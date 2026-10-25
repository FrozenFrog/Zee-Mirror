//go:build linux

package secretfs

import "golang.org/x/sys/unix"

func isRAMBacked(dir string) (bool, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(dir, &st); err != nil {
		return false, err
	}
	return st.Type == unix.TMPFS_MAGIC || st.Type == unix.RAMFS_MAGIC, nil
}
