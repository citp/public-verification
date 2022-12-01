module pvphm/nm

go 1.19

require github.com/citp/pvphm/bv v0.0.1

require (
	github.com/cloudflare/bn256 v0.0.0-20201110172847-66a4f6353b47 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de // indirect
	golang.org/x/sys v0.0.0-20200806125547-5acd03effb82 // indirect
)

replace github.com/citp/pvphm/bv => ../bv
