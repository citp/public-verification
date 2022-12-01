module pvphm/nm

go 1.13

require (
	github.com/citp/pvphm/bv v0.0.1
	github.com/cloudflare/bn256 v0.0.0-20201110172847-66a4f6353b47
	github.com/pkg/errors v0.9.1
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
)

replace github.com/citp/pvphm/bv => ../bv
