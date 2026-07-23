// sidebar-sse.js — listens for SSE events and triggers htmx sidebar refresh
// ponytail: single EventSource, reconnects automatically (htmx handles disconnect)

document.addEventListener('DOMContentLoaded', function () {
  var sidebar = document.getElementById('sidebar');
  if (!sidebar) return;

  var threadId = sidebar.getAttribute('data-thread-id');
  if (!threadId) return;

  var source = new EventSource('/v1/threads/' + threadId + '/events');

  source.addEventListener('turn.completed', function () {
    htmx.trigger('#sidebar', 'refresh');
    htmx.trigger('#timeline', 'refresh');
  });

  source.addEventListener('turn.failed', function () {
    htmx.trigger('#sidebar', 'refresh');
    htmx.trigger('#timeline', 'refresh');
  });

  source.onerror = function () {
    // ponytail: auto-reconnect is built into EventSource
  };
});
