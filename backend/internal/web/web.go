// Package web 内嵌前端构建产物，二进制单文件分发，不依赖文件系统。
// dist 由 frontend 的 pnpm build 生成（vite outDir 指向此处）。
package web

import "embed"

//go:embed all:dist
var Dist embed.FS
