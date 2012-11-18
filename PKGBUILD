pkgname=repoclean
pkgver=$(git log | wc -l)
pkgrel=1
pkgdesc='A tool to clean old versions of packages from an arch repo.'
arch=(x86_64 i686)
url=https://github.com/daaku/repoclean
source=(repoclean.go)
md5sums=($(md5sum ${source[*]} | sed -e 's/ .*//' | tr '\n' ' '))
license=(apache2)

package() {
  install -d $pkgdir/usr/bin
  go build -o $pkgdir/usr/bin/repoclean
}
