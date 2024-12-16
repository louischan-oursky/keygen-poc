package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/keygen-sh/keygen-go/v3"
	"github.com/keygen-sh/machineid"

	"github.com/louischan-oursky/keygen-poc/pkg/buildtimeconstant"
)

func cat(args []string) (err error) {
	// The spec says when there is no operands, stdin should be used.
	// See STDIN in https://pubs.opengroup.org/onlinepubs/9799919799/utilities/cat.html
	if len(args) == 0 {
		args = []string{"-"}
	}

	stdinUsed := false
	for _, arg := range args {
		if arg == "-" {
			if stdinUsed {
				// The spec says when the stdin has been used, the second occurrence of stdin is like reading from /dev/null.
				// We do not actually need to read from /dev/null.
				// See EXAMPLES in https://pubs.opengroup.org/onlinepubs/9799919799/utilities/cat.html
				continue
			}
			_, err = io.Copy(os.Stdout, os.Stdin)
			if err != nil {
				return
			}
			stdinUsed = true
		} else {
			var f *os.File
			f, err = os.Open(arg)
			if err != nil {
				return
			}
			defer f.Close()

			_, err = io.Copy(os.Stdout, f)
			if err != nil {
				return
			}
		}
	}

	return
}

func validateLicense(ctx context.Context, licenseKey string, fingerprint string, forceActivate bool) error {
	license, err := keygen.Validate(ctx, fingerprint)
	switch {
	case errors.Is(err, keygen.ErrLicenseNotActivated):
		_, err := license.Activate(ctx, fingerprint)
		switch {
		case errors.Is(err, keygen.ErrMachineLimitExceeded):
			if !forceActivate {
				fmt.Fprintf(os.Stderr, "machine limit exceeded. If you want to deactivate the previous machine, and activate this machine instead, add --force-activate\n")
				os.Exit(1)
			}
			machines, err := license.Machines(ctx)
			if err != nil {
				return err
			}
			for _, machine := range machines {
				err = machine.Deactivate(ctx)
				if err != nil {
					return err
				}
			}
			// Call itself recursively to retry, but set forceActivate=false so that we do not have infinite recursive.
			return validateLicense(ctx, licenseKey, fingerprint, false)
		case err != nil:
			return err
		}
	case errors.Is(err, keygen.ErrLicenseExpired):
		fmt.Fprintf(os.Stderr, "license is expired: %v\n", err)
		os.Exit(1)
	case err != nil:
		return err
	}

	return nil
}

func main() {
	keygen.APIURL = buildtimeconstant.KeygenAPIURL
	keygen.Account = buildtimeconstant.KeygenAccountID
	keygen.Product = buildtimeconstant.KeygenProductID

	var err error
	var licenseKey string
	var forceActivate bool
	var fingerprint string

	flag.StringVar(&licenseKey, "license", "", "The license key")
	flag.BoolVar(&forceActivate, "force-activate", false, "Deactivate any previous machine in order to activate this machine")
	flag.StringVar(&fingerprint, "fingerprint", "", "Override the default machine-id based fingerprint. This flag is for PoC purpose only.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "Build time configuration:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  KEYGEN_API_URL: %v\n", keygen.APIURL)
		fmt.Fprintf(flag.CommandLine.Output(), "  KEYGEN_ACCOUNT_ID: %v\n", keygen.Account)
		fmt.Fprintf(flag.CommandLine.Output(), "  KEYGEN_PRODUCT_ID: %v\n", keygen.Product)
	}

	flag.Parse()

	args := flag.Args()

	keygen.LicenseKey = licenseKey
	keygen.Logger = keygen.NewLogger(keygen.LogLevelNone)

	if fingerprint == "" {
		fingerprint, err = machineid.ProtectedID(keygen.Product)
		if err != nil {
			panic(err)
		}
	}

	ctx := context.Background()

	err = validateLicense(ctx, licenseKey, fingerprint, forceActivate)
	if err != nil {
		panic(err)
	}

	err = cat(args)
	if err != nil {
		panic(err)
	}
}
