// ============================================================================
// File: assets.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 04.01.2026
// Description: Embedded assets for WASM compatibility (images and model).
// Version: 1.0
//
// License: MIT
// Copyright 2025, School of Engineering and Architecture of Fribourg
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ============================================================================

package assets

import "embed"

// Embed images for stone rendering
//go:embed images/black-stone.png
var BlackStonePNG []byte

//go:embed images/white-stone.png
var WhiteStonePNG []byte

// Embed MLP model for AI player
//go:embed models/champion.json
var Models embed.FS
