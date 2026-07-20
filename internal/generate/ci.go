package generate

// GenerateCI returns the content of .github/workflows/pio_build.yml.
func GenerateCI() string {
	return `name: PlatformIO CI

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/cache@v4
        with:
          path: |
            ~/.cache/pip
            ~/.platformio/.cache
          key: ${{ runner.os }}-pio
      - uses: actions/setup-python@v5
        with:
          python-version: "3.11"
      - name: Install PlatformIO
        run: pip install platformio
      - name: Build
        run: pio run
`
}
