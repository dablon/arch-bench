// c-layered-http/src/transport: HTTP framing. The only file in the
// binary that knows about HTTP request/response, status codes, JSON
// encoding for the request body. It does NOT know about JWT.
#include "../service/token_service.h"
#include "../repo/jwt_repo.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#define MAX_REQ 8192

static int json_get_str(const char *json, const char *key, char *out, int outsz) {
    char needle[64];
    snprintf(needle, sizeof(needle), "\"%s\":\"", key);
    const char *p = strstr(json, needle);
    if (!p) return -1;
    p += strlen(needle);
    const char *e = strchr(p, '"');
    if (!e) return -1;
    int len = e - p;
    if (len >= outsz) len = outsz - 1;
    memcpy(out, p, len);
    out[len] = 0;
    return 0;
}

static void send_response(int fd, int code, const char *body) {
    const char *status = (code == 200) ? "OK" : (code == 400) ? "Bad Request" : (code == 401) ? "Unauthorized" : (code == 404) ? "Not Found" : (code == 405) ? "Method Not Allowed" : "Error";
    char resp[1024];
    int n = snprintf(resp, sizeof(resp),
        "HTTP/1.1 %d %s\r\nContent-Type: application/json\r\nContent-Length: %zu\r\nConnection: close\r\n\r\n%s",
        code, status, strlen(body), body);
    if (n > 0) { ssize_t _w = write(fd, resp, n); (void)_w; }
}

static int read_http(int fd, char *buf, int sz) {
    int total = 0;
    while (total < sz - 1) {
        int n = read(fd, buf + total, sz - 1 - total);
        if (n <= 0) break;
        total += n;
        buf[total] = 0;
        if (strstr(buf, "\r\n\r\n")) break;
    }
    return total;
}

static void handle(int fd) {
    char buf[MAX_REQ];
    int n = read_http(fd, buf, sizeof(buf));
    if (n <= 0) { close(fd); return; }
    char method[16] = {0}, path[256] = {0};
    sscanf(buf, "%15s %255s", method, path);
    if (strcmp(path, "/health") == 0) {
        send_response(fd, 200, "{\"status\":\"ok\"}");
    } else if (strcmp(path, "/verify") == 0) {
        if (strcmp(method, "POST") != 0) {
            send_response(fd, 405, "{\"valid\":false,\"code\":\"ERR_METHOD\"}");
        } else {
            char *body = strstr(buf, "\r\n\r\n");
            if (!body) { send_response(fd, 400, "{\"valid\":false,\"code\":\"ERR_BAD_REQUEST\"}"); close(fd); return; }
            body += 4;
            char token[1024] = {0};
            if (json_get_str(body, "token", token, sizeof(token)) != 0 || !*token) {
                send_response(fd, 400, "{\"valid\":false,\"code\":\"ERR_BAD_REQUEST\"}");
            } else {
                ServiceResult res;
                service_verify(token, &res);
                if (res.valid) {
                    char resp[256];
                    snprintf(resp, sizeof(resp), "{\"valid\":true,\"subject\":\"%s\",\"code\":\"OK\"}", res.subject);
                    send_response(fd, 200, resp);
                } else if (res.code == CODE_BAD_REQUEST) {
                    send_response(fd, 400, "{\"valid\":false,\"code\":\"ERR_BAD_REQUEST\"}");
                } else {
                    send_response(fd, 401, "{\"valid\":false,\"code\":\"ERR_INVALID_TOKEN\"}");
                }
            }
        }
    } else {
        send_response(fd, 404, "{\"valid\":false,\"code\":\"ERR_NOT_FOUND\"}");
    }
    close(fd);
}

int main(int argc, char **argv) {
    (void)argc; (void)argv;
    const char *secret = getenv("JWT_SECRET");
    const char *addr_env = getenv("LISTEN_ADDR");
    if (!secret) { fprintf(stderr, "JWT_SECRET required\n"); return 1; }
    int port = 9001;
    if (addr_env) {
        char *colon = strrchr(addr_env, ':');
        if (colon) port = atoi(colon + 1);
    }
    g_secret_len = strlen(secret);
    if (g_secret_len >= AB_SECRET_MAX) { fprintf(stderr, "secret too long\n"); return 1; }
    memcpy(g_secret, secret, g_secret_len);

    int srv = socket(AF_INET, SOCK_STREAM, 0);
    int opt = 1;
    setsockopt(srv, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
    struct sockaddr_in sa = {0};
    sa.sin_family = AF_INET;
    sa.sin_addr.s_addr = htonl(INADDR_LOOPBACK);
    sa.sin_port = htons(port);
    if (bind(srv, (struct sockaddr *)&sa, sizeof(sa)) < 0) { perror("bind"); return 1; }
    listen(srv, 64);
    fprintf(stderr, "c-layered-http listening on 127.0.0.1:%d\n", port);
    signal(SIGCHLD, SIG_IGN);
    while (1) {
        int fd = accept(srv, NULL, NULL);
        if (fd < 0) continue;
        if (fork() == 0) { close(srv); handle(fd); return 0; }
        close(fd);
    }
    return 0;
}
