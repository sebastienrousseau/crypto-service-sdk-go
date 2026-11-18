# SPDX-License-Identifier: Apache-2.0 OR MIT
# SPDX-FileCopyrightText: 2026 Sebastien Rousseau
#
# Rendered by `make demo` (vhs, https://github.com/charmbracelet/vhs)
# from the repository root, against a crypto-server already running on
# loopback at CRYPTO_SERVICE_URL with CRYPTO_SERVICE_API_KEY set. Hidden
# from the recording: building the examples into bin/. On screen: the
# hash, aead, sign and mlkem examples. Only loopback is used, and the
# key in the aead example is a fixed synthetic test value.
Output .github/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"

Hide
Type `go build -o bin/ ./examples/... && export PS1='$ ' && clear`
Enter
Wait@120s /^\$$/
Show

Type "bin/hash"
Sleep 500ms
Enter
Wait@30s /^\$$/
Sleep 2s

Type "bin/aead"
Sleep 500ms
Enter
Wait@30s /^\$$/
Sleep 2s

Type "bin/sign"
Sleep 500ms
Enter
Wait@30s /^\$$/
Sleep 2s

Type "bin/mlkem"
Sleep 500ms
Enter
Wait@30s /^\$$/
Sleep 4s
