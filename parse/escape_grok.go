package parse

// // escapeGrokArgs escapes unescaped Grok patterns in an HCL file
//
//	func escapeGrokArgs(fileData []byte, filePath string) []byte {
//		for {
//			var changed bool
//			fileData, changed = doEscapeGrokArgs(fileData, filePath)
//			if !changed {
//				break
//			}
//		}
//		return fileData
//	}
// func doEscapeGrokArgs(fileData []byte, filePath string) ([]byte, bool) {
// 	// Parse HCL file without caching
// 	file, diags := hclsyntax.ParseConfig(fileData, filePath, hcl.Pos{Byte: 0, Line: 1, Column: 1})

// 	// Return original data if no errors or failed parsing
// 	if !diags.HasErrors() || file == nil {
// 		return fileData, false
// 	}

// 	type replacement struct {
// 		start int
// 		end   int
// 		value string
// 	}

// 	var replacements []replacement

// 	// Iterate over diagnostics to find Grok pattern errors
// 	for _, diag := range diags {
// 		if diag.Summary == "Invalid template control keyword" || diag.Detail == "Expected the start of an expression, but found an invalid expression token." {

// 			a := getAttributeForRange(file.Body.(*hclsyntax.Body), diag.Subject)
// 			if a == nil {
// 				continue
// 			}

// 			// we only do this escaping if the attribute is a 'grok' function call
// 			f, ok := a.Expr.(*hclsyntax.FunctionCallExpr)
// 			if !ok || f.Name != "grok" {
// 				continue
// 			}
// 			if len(f.Args) == 0 {
// 				continue
// 			}
// 			arg, ok := f.Args[0].(*hclsyntax.LiteralValueExpr)
// 			if !ok {
// 				continue
// 			}

// 			startByte := arg.SrcRange.Start.Byte
// 			// The end byte will not be set as the parse of the arg failed, so just take the while line
// 			// Find the end of the current line
// 			// Default to end of file
// 			endByte := len(fileData) - 1
// 			for i := startByte + 1; i < len(fileData); i++ {
// 				if fileData[i] == '\n' {
// 					endByte = i - 1
// 					break
// 				}
// 			}
// 			// Extract the value to be escaped
// 			hclVal := fileData[startByte:endByte]

// 			// Efficiently escape "%{" while keeping existing "%%{" unchanged
// 			escapedAttr := escapeGrokCapturePattern(string(hclVal))

// 			// add to list of replacements
// 			//
// 			replacements = append(replacements, replacement{start: startByte, end: endByte, value: escapedAttr})
// 		}
// 	}

// 	for i := len(replacements) - 1; i >= 0; i-- {
// 		// Perform replacements in reverse order to avoid changing the indices
// 		replacement := replacements[i]
// 		fileData = append(fileData[:replacement.start], append([]byte(replacement.value), fileData[replacement.end:]...)...)
// 	}

// 	return fileData, len(replacements) > 0
// }

// escapeGrokCapturePattern ensures "%{" is escaped as "%%{" but does NOT double-escape existing "%%{"
// escapeGrokCapturePattern escapes "%{" as "%%{" but does NOT double-escape existing "%%{"
// func escapeGrokCapturePattern(input string) string {
// 	var sb strings.Builder
// 	n := len(input)

// 	for i := 0; i < n; i++ {
// 		if input[i] == '%' && i+1 < n && input[i+1] == '{' {
// 			// If it's already "%%{", keep it as is
// 			if i > 0 && input[i-1] == '%' {
// 				sb.WriteString("%{") // Keep it unchanged
// 			} else {
// 				sb.WriteString("%%{") // Escape "%{" to "%%{"
// 			}
// 			i++ // Skip '{' since we already processed it
// 		} else {
// 			sb.WriteByte(input[i])
// 		}
// 	}

// 	return sb.String()
// }

//
//// escapeGrokProperties escapes template expressions in any properties specified by escapeGrokProperties
//func escapeGrokArgs(fileData []byte, filePath string) []byte {
//
//	// do an initial parse of the file
//	// NOTE: we call hclsyntax.ParseConfig directly - do not want to use the parser as we do not want to cache the result
//	file, diags := hclsyntax.ParseConfig(fileData, filePath, hcl.Pos{Byte: 0, Line: 1, Column: 1})
//
//	// if there are no errors - or idf we failed to even parse the file - we are done
//	if !diags.HasErrors() || file == nil {
//		return fileData
//	}
//
//	//syntaxBody := file.Body.(*hclsyntax.Body)
//	for _, diag := range diags {
//		if diag.Summary == "Invalid template control keyword" || diag.Detail == "Expected the start of an expression, but found an invalid expression token." {
//
//			//attr := getAttributeForRange(syntaxBody, diag.Subject)
//			//if attr == nil {
//			//	continue
//			//}
//
//			startByte := diag.Subject.Start.Byte
//
//			// find the end newline
//			endByte := 0
//
//			for i := startByte + 1; i < len(fileData); i++ {
//				if fileData[i] == '\n' || fileData[i] == '\r' {
//					endByte = i - 2
//					break
//				}
//			}
//			if endByte == 0 {
//				endByte = len(fileData) - 1
//			}
//
//			hclVal := fileData[startByte:endByte]
//
//			// escape %{
//			// Regex to match "%{" only if NOT preceded by "%"
//			re := regexp.MustCompile(`([^%]|^)%{`)
//
//			// Replace "%{" with "%%{" (but only if not already escaped) while preserving the preceding character
//			escapedAttr := e
//
//			start := fileData[:startByte]
//			middle := []byte(escapedAttr)
//			end := fileData[endByte:]
//			// rebuild the fileData
//			fileData = append(start, append(middle, end...)...)
//
//		}
//	}
//	return fileData
//}
//
//// Approach 2: Manual iteration with strings.Builder
//func escapeGrokPatternManual(input string) string {
//	var sb strings.Builder
//	n := len(input)
//
//	for i := 0; i < n; i++ {
//		if input[i] == '%' && i+1 < n && input[i+1] == '{' {
//			// Check if the previous character is also '%'
//			if i > 0 && input[i-1] == '%' {
//				sb.WriteByte('%') // Keep "%%{" unchanged
//			} else {
//				sb.WriteString("%%") // Escape "%{" to "%%{"
//			}
//		}
//		sb.WriteByte(input[i])
//	}
//	return sb.String()
//}

// func getAttributeForRange(syntaxBody *hclsyntax.Body, subject *hcl.Range) *hclsyntax.Attribute {
// 	for _, attribute := range syntaxBody.Attributes {
// 		if attribute.Expr.Range().Start.Byte <= subject.Start.Byte && attribute.Expr.Range().End.Byte >= subject.End.Byte {
// 			return attribute
// 		}
// 	}
// 	for _, block := range syntaxBody.Blocks {
// 		attr := getAttributeForRange(block.Body, subject)
// 		if attr != nil {
// 			return attr
// 		}
// 	}

// 	return nil

// }
