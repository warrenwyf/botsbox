#include <gio/gio.h>

typedef struct { void *fn; } callback_wrapper;

callback_wrapper *make_callback_wrapper();

void free_callback_wrapper(callback_wrapper *p);

extern void goGAsyncReadyCallback(void *result, gpointer userData);

void my_g_async_ready_callback(GObject *source_object, GAsyncResult *res,
                               gpointer user_data);