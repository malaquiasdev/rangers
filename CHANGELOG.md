# Changelog

## 0.1.0 (2026-11-22)


### Features

* add the learn command in Go ([6f046be](https://github.com/malaquiasdev/rangers/commit/6f046be140c10fa362ddb68e317fbcaa23baadee))
* **agents:** add a reminder as the second channel for what the human must answer ([8839a28](https://github.com/malaquiasdev/rangers/commit/8839a28917d6d450352c954edd08dfd65b1a182e))
* **assets:** add a pixel art toolkit with ramps, shading and dithering ([b9da6d1](https://github.com/malaquiasdev/rangers/commit/b9da6d117fbf3ee20afdd45aa6b68a7ef09f1cd2))
* **assets:** add a shape and lighting toolkit for the pixel art ([3ab525b](https://github.com/malaquiasdev/rangers/commit/3ab525b7957054428a223acec62d92d4ca9ae1fa))
* **assets:** compose the hero poster and the wide banner ([f66b63a](https://github.com/malaquiasdev/rangers/commit/f66b63a65f2ae7845e82ace4dc8c1a748bef5e28))
* **assets:** draw full-body sprites for the 11 agents ([e2f0b79](https://github.com/malaquiasdev/rangers/commit/e2f0b796ca2fef4cfb04fc8b4d488a17f5d982fc))
* **assets:** drop the heroes from the poster and the banner and build a stage ([665d78c](https://github.com/malaquiasdev/rangers/commit/665d78cbc6169ad90d9738cd0e042721db94eb13))
* **assets:** redo the banner as a layered 16-bit title screen ([a8e2c91](https://github.com/malaquiasdev/rangers/commit/a8e2c91b8d71873893132088b96f57fb8cb1f37f))
* **assets:** redraw the six hero sprites as human-proportioned figures ([7cae6df](https://github.com/malaquiasdev/rangers/commit/7cae6df1dc616b34badb3e69b6c5711987926978))
* **bin:** add the audit helper ([070b4a7](https://github.com/malaquiasdev/rangers/commit/070b4a77c5fd964b98c2dde7b4ccb7e0f95ef4a8))
* **deps:** pin grill-with-docs, domain-modeling and the two Obsidian skills ([ef211da](https://github.com/malaquiasdev/rangers/commit/ef211da00f324c37ae0d369eb18fb9e6660f98c2))
* **gate:** bind the human's yes to the head SHA and add gate ask ([b3f76a2](https://github.com/malaquiasdev/rangers/commit/b3f76a29a345d2cd734492f51e7fa12ca24ba3bd))
* **gate:** check the pull request state and stop at production ([e1712ba](https://github.com/malaquiasdev/rangers/commit/e1712ba2e5c1d17721e548a2267769b7066c8c8c))
* **gate:** refuse pull requests whose checks are not all green ([7df6fb5](https://github.com/malaquiasdev/rangers/commit/7df6fb5a17d37e608c7742ccc607deef3a9263fc))
* **gate:** squash pinned to the head SHA with --merge and write the audit line ([b9ce43d](https://github.com/malaquiasdev/rangers/commit/b9ce43df7158f947531719ad40a2f96d502a4a63))
* **gate:** wire rangers gate with exit codes 0, 1 and 2 ([4697034](https://github.com/malaquiasdev/rangers/commit/4697034eb6cfc6a860b708672435eb0d576f36c6))
* **install:** add rangers install --check and replace verify.sh ([d1603b7](https://github.com/malaquiasdev/rangers/commit/d1603b7f741fe932f1371a294f6b1a60a4c0cb75))
* **install:** install the missing pieces, link agents, rules and the skill ([722f8c5](https://github.com/malaquiasdev/rangers/commit/722f8c50c0d26571b3822e7bf7e28dd59cb15a7c))
* **install:** let the files port tell links from regular files ([5538c7d](https://github.com/malaquiasdev/rangers/commit/5538c7d2c2fd0be1636f268800bcf00a93cf90f9))
* **install:** link bin files into the claude dir ([4f695d8](https://github.com/malaquiasdev/rangers/commit/4f695d824aea34a8b4e6c395a4f38583373641d6))
* **install:** link every skills folder into the claude dir ([a8d5773](https://github.com/malaquiasdev/rangers/commit/a8d5773f3626b931f3570735585b1f3ad51049ec))
* **install:** report each skill link in --check ([9c47538](https://github.com/malaquiasdev/rangers/commit/9c475380cc3adc6bdef76fdaac432ef3de800d3d))
* **install:** report the bin links in --check ([dc88b96](https://github.com/malaquiasdev/rangers/commit/dc88b9695064b8753b47bea23350a17e518e0291))
* **learn:** add a daily dollar cap that replaces a limit on the number of rounds ([bc16e78](https://github.com/malaquiasdev/rangers/commit/bc16e7867205976cc6fa3080a7a7262e8c3cf855))
* **learn:** add a notes pass that writes decisions and results to the brain with Sonnet ([c2f363d](https://github.com/malaquiasdev/rangers/commit/c2f363d30cbc2da5b5963f4736d566f93c2f3f49))
* **learn:** add the call timeout setting ([6dfefcc](https://github.com/malaquiasdev/rangers/commit/6dfefcca1e2439038100f597cc8d79f2fdc144bc))
* **learn:** file the origin of each rule in the private notes ([be36fb3](https://github.com/malaquiasdev/rangers/commit/be36fb3fc17884bd2e32e668aa499fea25a61759))
* **rules:** add four rules learned from the human's corrections ([ee37360](https://github.com/malaquiasdev/rangers/commit/ee37360655609a845ad8e330cb08377dd20982b4))
* **schedule:** replace the shell schedule script with the rangers schedule command ([88d1749](https://github.com/malaquiasdev/rangers/commit/88d17492616d746e667cfbdd21283931c3d486aa))
* **scripts:** add a launchd schedule for the learn command every two hours ([3a14459](https://github.com/malaquiasdev/rangers/commit/3a14459c3fbd88b49628aa30e0c2ce3b05c910cc))
* **skills:** add claude-sessions and pr-description ([7cddc21](https://github.com/malaquiasdev/rangers/commit/7cddc2120386add29ad4cdd91f0b6a04ea829860))
* **transcript:** read whole conversations with both speakers ([0ee5b7b](https://github.com/malaquiasdev/rangers/commit/0ee5b7b437db766e1f04fda3fd140c78816980dd))


### Bug Fixes

* **gate:** refuse a pull request with no checks instead of failing ([7f7912c](https://github.com/malaquiasdev/rangers/commit/7f7912c64137e6f96c6a8128760f884f801c5a25))
* **learn:** a failed reminder is a warning and a short timeout keeps the job from hanging ([55667fd](https://github.com/malaquiasdev/rangers/commit/55667fd568d16807f75c986d857c58f76cf53d25))
* **learn:** do not wait for the output of a claude child after the kill ([e2b26ef](https://github.com/malaquiasdev/rangers/commit/e2b26eff45ebb7d3bff1cd609b999b3829cc3a7d))
* **learn:** keep origins out of the rules and the commits ([9ec3bfb](https://github.com/malaquiasdev/rangers/commit/9ec3bfb3bf5fd3e95df3444a857b0c479514638a))
* **learn:** skip the notes pass when the brain folder is not accessible ([1464c04](https://github.com/malaquiasdev/rangers/commit/1464c046492935039e655f40e50d9d3f39a35d66))
* **learn:** stop a claude call that outlives the configured timeout ([c91ee26](https://github.com/malaquiasdev/rangers/commit/c91ee264332bed5fcc3168786bd39452c48d7df7))
* **rules:** remove the origin quotes from the rules file ([978c24e](https://github.com/malaquiasdev/rangers/commit/978c24e65f317f156b2f60434afdaa8445f297a6))
* **scripts:** put the Go tools in the PATH of the scheduled job ([96beb61](https://github.com/malaquiasdev/rangers/commit/96beb61d6cc25afb7c8d09c5cd0945cec7c66a15))
