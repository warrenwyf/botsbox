#include "gasyncreadycallback.go.h"

callback_wrapper *make_callback_wrapper() {
  callback_wrapper *p = (callback_wrapper *)malloc(sizeof(callback_wrapper));
  return p;
}

void free_callback_wrapper(callback_wrapper *p) {
  p->fn = NULL;
  free(p);
}

void my_g_async_ready_callback(GObject *source_object, GAsyncResult *res,
                               gpointer user_data) {
  goGAsyncReadyCallback(res, user_data);
}