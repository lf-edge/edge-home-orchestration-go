/*******************************************************************************
 * Copyright 2020 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/

/* Included Files. */
#include <stdio.h>
#include <orchestration.h>
#include <pthread.h>
#include <unistd.h>
#include <getopt.h>
#include <string.h>

/* Private Data. */
/* All static data definitions appear here. */
RequestServiceInfo rsi = {
    .ExecutionType  = "native",
    .ExeCmd         = "ls"
};

ResponseService rs;

char str[1024];

void usage(const char *path)
{
    const char *basename = strrchr(path, '/');
    basename = basename ? basename + 1 : path;

    printf ("usage: %s [OPTION]\n", basename);
    printf ("  -h, --help\t\t"
                 "Print this help and exit\n");
    printf ("  -s, --secure[=true]\t"
                 "Edge Orchestration will be run in secure mode\n");
    printf ("  -m, --mnedc=STRING\t"
                 "Edge Orchestration will be run as MNEDC server/client\n");
}

int main(int argc, char *argv[]) {

    int secure = 0;
    int mndec = 0;
    int opt;

    struct option longopts[] = {
        { "help", no_argument, NULL, 'h' },
        { "mndec", required_argument, NULL, 'm' },
        { "secure", optional_argument, NULL, 's' },
        { 0 }
    };

    while ((opt = getopt_long (argc, argv, "hm:s::", longopts, 0)) != -1) {
        switch (opt) {
            case 'h':
                usage(argv[0]);
                return 0;
            case 'm':
                if (strcmp("server", optarg) == 0)
                    mndec = 1;
                if (strcmp("client", optarg) == 0)
                    mndec = 2;
                break;
            case 's':
                if (optarg != 0) {
                    if (strcmp("true", optarg) == 0)
                        secure = 1;
                    else
                        break;
                }
                secure = 1;

                break;
            case '?':
                usage(argv[0]);
                return 1;
            default:
                break;
            }
    }

    argc = 1;
    argv[1] = 0;

    if (OrchestrationInit(secure, mndec) != 0) {
        return -1;
    }

    rs = OrchestrationRequestService("ls", 1, "bash", &rsi, 1);

    sprintf(str, "Message: %s", rs.Message);
    PrintLog(str);
    sprintf(str, "ServiceName: %s", rs.ServiceName);
    PrintLog(str);
    sprintf(str, "ExecutionType: %s", rs.RemoteTargetInfo.ExecutionType);
    PrintLog(str);
    sprintf(str, "Target: %s", rs.RemoteTargetInfo.Target);
    PrintLog(str);

    for (;;) {
        sleep(1);
    }

    return 0;
}

