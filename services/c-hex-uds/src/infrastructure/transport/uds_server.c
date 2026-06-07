// c-hex-uds/src/infrastructure/transport/uds_server.c
// INCOMING adapter. UDS line protocol, composition root.
#include "../../domain/usecase.h"
#include "../../domain/result.h"
#include "../jwt/jwt_adapter.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include <sys/stat.h>
#include <sys/socket.h>
#include <sys/un.h>

#define MAX_LINE 2048

static void handle(int fd, const TokenVerifierPort *verifier_port) {
    char buf[MAX_LINE];
    while (1) {
        ssize_t n = read(fd, buf, sizeof(buf) - 1);
        if (n <= 0) break;
        buf[n] = 0;
        char *nl = strchr(buf, '\n');
        if (nl) *nl = 0;
        char okbuf[256];
        if (strncmp(buf, "VERIFY ", 7) != 0) {
            ssize_t w = write(fd, "ERR BAD_REQUEST\n", 16); (void)w;
            continue;
        }
        VerificationResult res;
        usecase_verify(verifier_port, buf + 7, &res);
        if (res.valid) {
            snprintf(okbuf, sizeof(okbuf), "OK %s\n", res.subject);
        } else if (res.code == DCODE_BAD_REQUEST) {
            snprintf(okbuf, sizeof(okbuf), "ERR BAD_REQUEST\n");
        } else {
            snprintf(okbuf, sizeof(okbuf), "ERR INVALID_TOKEN\n");
        }
        ssize_t w = write(fd, okbuf, strlen(okbuf)); (void)w;
    }
    close(fd);
}

int main(int argc, char **argv) {
    (void)argc; (void)argv;
    const char *secret = getenv("JWT_SECRET");
    const char *path_env = getenv("UDS_PATH");
    if (!secret) { fprintf(stderr, "JWT_SECRET required\n"); return 1; }
    const char *path = path_env ? path_env : "/tmp/c-hex-uds.sock";
    g_secret_len = strlen(secret);
    if (g_secret_len >= AB_SECRET_MAX) { fprintf(stderr, "secret too long\n"); return 1; }
    memcpy(g_secret, secret, g_secret_len);

    TokenVerifierPort verifier_port;
    jwt_adapter_port(&verifier_port);

    unlink(path);
    int srv = socket(AF_UNIX, SOCK_STREAM, 0);
    struct sockaddr_un sa = {0};
    sa.sun_family = AF_UNIX;
    strncpy(sa.sun_path, path, sizeof(sa.sun_path) - 1);
    if (bind(srv, (struct sockaddr *)&sa, sizeof(sa)) < 0) { perror("bind"); return 1; }
    listen(srv, 64);
    chmod(path, 0660);
    fprintf(stderr, "c-hex-uds listening on %s\n", path);
    signal(SIGCHLD, SIG_IGN);
    while (1) {
        int fd = accept(srv, NULL, NULL);
        if (fd < 0) continue;
        if (fork() == 0) { close(srv); handle(fd, &verifier_port); return 0; }
        close(fd);
    }
    return 0;
}
