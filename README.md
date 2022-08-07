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
      --pre=PRE          the prerelease suffix
      --remote="origin"  the git remote
      --repo=.           the directory of git repository
  -y, --yes              answer yes to all questions
      --debug            enable debug mode

Args:
  [<tag>]  the tag to create 
```

## Example
![example](https://user-images.githubusercontent.com/761734/183281408-479dd4d8-c761-48a3-af5f-965561494d34.gif)