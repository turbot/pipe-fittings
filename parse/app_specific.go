package parse

// ModDecoderFunc is the appspecific constructor function used to construct a decoder which knows how to
// decode resources for the app in question
var ModDecoderFunc func(...DecoderOption) Decoder
