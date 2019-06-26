
/*
$ gcc -fPIC -c myscoring.c && gcc -shared -o libmyscoring.so.1.0.1 myscoring.o -lm
$ ln -rs libmyscoring.so.1.0.1 libmyscoring.so
*/

#include <math.h>
#include <assert.h>
#include <stdio.h>

#define CNT 6

/*
features :
-. there is using moving average, but it is not written code at service_provider.cpp
-. bandwidth is not meaning Mbps
*/

//network score mmDiscovery/service_provider.cpp
static double networkScore(double n)
{
  return 1 / (8770 * pow(n, -0.9));
}

//cpu score mmDiscovery/service_provider.cpp
static double cpuScore(double freq, double usage, double count)
{
  return ((1 / (5.66 * pow(freq, -0.66))) +
          (1 / (3.22 * pow(usage, -0.241))) +
          (1 / (4 * pow(count, -0.3)))) /
         3;
}

//render score mmDiscovery/service_provider.cpp
//https://github.com/Samsung/Castanets/blob/castanets_63/service_discovery_manager/Component/mmDiscovery/monitor_client.cpp
static double renderingScore(double r)
{
  return (r < 0) ? 0 : 0.77 * pow(r, -0.43);
}

//============== INTERFACE API ==============
// double myscoring(double (*getResource)(const char *))
double myscoring(int (*getResource)(const char *, double * out))
{
  printf("myscoring\n");
  int ret;
  double score;
  double networkBandwidth;
  double cpuFreq, cpuUsage, cpuCount;

  score = 0.0;
  ret = getResource("network/bandwidth", &networkBandwidth);
  if(ret != 0) {
    return 0.0;
  } else {
    score += networkScore(networkBandwidth);
  }

  ret = getResource("cpu/freq", &cpuFreq);
  if(ret != 0) {
    return 0.0;
  }

  ret = getResource("cpu/usage", &cpuUsage);
  if(ret != 0) {
    return 0.0;
  }

  ret = getResource("cpu/count", &cpuCount);
  if(ret != 0) {
    return 0.0;
  }
  score += cpuScore(cpuFreq, cpuUsage, cpuCount);
  score /= 2;
  // score += renderingScore(getResource("network/rtt"));

  return score;
}

// #define CNT 6

// double myscoring2(double (*getResource)(const char *))
// {

//   printf("myscoring\n");

//   double score;
//   const char *resourceNames[CNT] = {"cpu/usage", "cpu/count", "memory/free", "memory/available", "network/mbps", "network/bandwidth"};
//   double W[CNT] = {1.48271, 4.125421, 5.3381723, 9.194717234, 2.323, 1.123};
//   double resourceValues[CNT];

//   // double someResource;
//   // someResource = getResource("some/usage");
//   // assert(isnan(someResource));

//   for (int i = 0; i < CNT; i++)
//   {
//     resourceValues[i] = getResource(resourceNames[i]);
//     printf("resourceNames : %s %f\n", resourceNames[i], resourceValues[i]);
//   }

//   score = 0.0;
//   for (int i = 0; i < CNT; i++)
//   {
//     score += resourceValues[i] * W[i];
//   }

//   return score;
// }

// #define MY_SCORING_3_CNT 4
// double myscoring3(double (*getResource)(const char *))
// {

//   // printf("myscoring\n");

//   double score;
//   const char *resourceNames[MY_SCORING_3_CNT] = {"cpu/usage", "cpu/count", "memory/free", "memory/available"};
//   double W[MY_SCORING_3_CNT] = {1.48271, 4.125421, 5.3381723, 9.194717234};
//   double resourceValues[MY_SCORING_3_CNT];

//   // double someResource;
//   // someResource = getResource("some/usage");
//   // assert(isnan(someResource));

//   for (int i = 0; i < MY_SCORING_3_CNT; i++)
//   {
//     resourceValues[i] = getResource(resourceNames[i]);
//     // printf("resourceNames : %s %f\n", resourceNames[i], resourceValues[i]);
//   }

//   score = 0.0;
//   for (int i = 0; i < MY_SCORING_3_CNT; i++)
//   {
//     score += resourceValues[i] * W[i];
//   }

//   return score;
// }
