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

/* Private Data. */
/* All static data definitions appear here. */
RequestServiceInfo rsi = { 
  .ExecutionType  = "native",
  .ExeCmd         = "ls"
};

ResponseService rs;

char str[1024];

int main(int argc, char *argv[])
{

	if (OrchestrationInit() != 0) {
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

  for(;;) {
    sleep(1);
  }

  return 0;
}

