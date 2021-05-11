#include <stdio.h>
#include <string.h>
#include "callback.h"
#include "_cgo_export.h"

/* for OPT_HEADERFUNCTION */
size_t header_function( char *ptr, size_t size, size_t nmemb, void *ctx) {
	return goCallHeaderFunction(ptr, size*nmemb, ctx);
}

void *return_header_function() {
    return (void *)&header_function;
}


/* for OPT_WRITEFUNCTION */
size_t write_function( char *ptr, size_t size, size_t nmemb, void *ctx) {
	return goCallWriteFunction(ptr, size*nmemb, ctx);
}

void *return_write_function() {
    return (void *)&write_function;
}

/* for OPT_READFUNCTION */
size_t read_function( char *ptr, size_t size, size_t nmemb, void *ctx) {
	return goCallReadFunction(ptr, size*nmemb, ctx);
}

void *return_read_function() {
    return (void *)&read_function;
}


/* for OPT_PROGRESSFUNCTION */
int progress_function(void *ctx, double dltotal, double dlnow, double ultotal, double ulnow) {
	return goCallProgressFunction(dltotal, dlnow, ultotal, ulnow, ctx);
}

void *return_progress_function() {
    return (void *)progress_function;
}

/* for OPT_DEBUGFUNCTION */
int debug_function(CURL *handle, curl_infotype type, char *data, size_t size, void *ctx) {
	switch (type) {
	case CURLINFO_TEXT:
	case CURLINFO_HEADER_IN:
	case CURLINFO_HEADER_OUT:
	case CURLINFO_END:
		return goCallDebugFunction(data, size, ctx);
        }
}

void *return_debug_function() {
    return (void *)&debug_function;
}
