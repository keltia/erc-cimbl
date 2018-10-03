package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/proglottis/gpgme"
)

// decryptFiles returns the path name of the decrypted file
func decryptFile(ctx *Context, file string) (string, error) {
	// Carefully open the box
	fh, err := os.Open(file)
	if err != nil {
		return "", errors.Wrap(err, "decryptFile/Open")
	}
	defer fh.Close()

	// Do the decryption thing
	plain, err := gpgme.Decrypt(fh)
	if err != nil {
		return "", errors.Wrap(err, "Decrypt")
	}
	defer plain.Close()

	// Save "plain" text
	base := filepath.Base(file)
	ext := filepath.Ext(base)
	zipname := strings.Replace(base, ext, "", 1)

	plainfile := filepath.Join(ctx.tempdir.Cwd(), zipname)

	verbose("Decrypting %s as %s", file, plainfile)

	dfh, err := os.Create(plainfile)
	if err != nil {
		return "", errors.Wrap(err, "decryptFile/Create")
	}
	defer dfh.Close()

	_, err = io.Copy(dfh, plain)
	if err != nil {
		return "", errors.Wrap(err, "decryptFile/Copy")
	}

	return plainfile, nil
}
