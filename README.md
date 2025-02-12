# gittag

gittag は Semantic Versioning のルールに従って、 git のタグを作成してプッシュするためのツールです。
次のバージョンの候補を `git ls-remote --tags` の内容に基づいて提示してくれます。

## Install

```bash
$ go install github.com/kazz187/gittag@latest
```

## Usage

```bash
$ gittag

$ gittag --help
usage: gittag [<flags>] [<tag>]

Semantic versioning tagging tool

Flags:
      --help             Show context-sensitive help (also try --help-long and --help-man).
  -s, --segment=SEGMENT  the segment to increment
  -p, --pre=PRE          the prerelease suffix
      --remote="origin"  the git remote
      --repo=.           the directory of git repository
  -y, --yes              answer yes to all questions
      --debug            enable debug mode

Args:
  [<tag>]  the tag to create
```

環境変数でSSHの方法を変更できます。

- `GITTAG_SSH_AUTH_SOCK`: SSH に利用する ssh-agent のソケットファイルのパスを指定できます。(`SSH_AUTH_SOCK` が指定されている場合、指定しなくても使われます。)
- `GITTAG_SSH_KEY`: SSH に利用する秘密鍵のパスを指定できます。 (ssh-agent が利用できない状態でこの変数を指定しない場合、`~/.ssh/id_rsa` が利用されます。)
- `GITTAG_SSH_PASSPHRASE`: SSH に利用する秘密鍵のパスフレーズを指定できます。

## Example
![example](https://user-images.githubusercontent.com/761734/183281408-479dd4d8-c761-48a3-af5f-965561494d34.gif)