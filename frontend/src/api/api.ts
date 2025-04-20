import {
  AppointmentsApi,
  AuthApi,
  ConditionsApi,
  DoctorsApi,
  ErrorDetail,
  FetchError,
  MedicalHistoryApi,
  PatientsApi,
  ResourcesApi,
  ResponseError,
} from './generated';
import { Configuration, FetchAPI } from './generated';

export interface Api {
  auth: AuthApi;
  appointments: AppointmentsApi;
  conditions: ConditionsApi;
  doctors: DoctorsApi;
  patients: PatientsApi;
  medicalHistory: MedicalHistoryApi;
  resources: ResourcesApi;
}

const fetchApi: FetchAPI = async (input, init): Promise<Response> => {
  if (!init) init = {};
  init.signal = AbortSignal.timeout(5000);

  const start = Date.now();
  const result = fetch(input, init);
  const end = Date.now();

  const url = input instanceof Request ? input.url : input;
  console.log(`[FETCH] ${url} took ${end - start}ms to execute`);

  return result;
};

export function newApi(apiBase: string): Api {
  const config = new Configuration({ basePath: apiBase, fetchApi });
  const api = {
    auth: new AuthApi(config),
    appointments: new AppointmentsApi(config),
    conditions: new ConditionsApi(config),
    doctors: new DoctorsApi(config),
    patients: new PatientsApi(config),
    medicalHistory: new MedicalHistoryApi(config),
    resources: new ResourcesApi(config),
  };

  return new Proxy(api, apiProxyHandler);
}

export class ApiError extends Error {
  override name: 'ApiError' = 'ApiError' as const;
  constructor(public errDetail: ErrorDetail) {
    super(errDetail.detail);
  }
}

const methodProxyHandler: ProxyHandler<any> = {
  get(target, propKey, receiver) {
    const originalMethod = Reflect.get(target, propKey, receiver);

    // If the property is a function (an API method), wrap it
    if (typeof originalMethod === 'function') {
      return async (...args: any[]) => {
        const methodName = String(propKey);
        try {
          const result = await originalMethod.apply(target, args);
          return result;
        } catch (err) {
          if (err instanceof ResponseError) {
            const errDetail = await err.response.json();
            throw new ApiError(errDetail);
          } else if (err instanceof FetchError) {
            throw err.cause;
          } else {
            console.error(
              `[PROXY] Caught non-ResponseError in ${target.constructor.name}.${methodName}.`,
            );
            throw err;
          }
        }
      };
    }

    return originalMethod;
  },
};

const apiProxyHandler: ProxyHandler<Api> = {
  get(target, propKey, receiver) {
    const originalApiClient = Reflect.get(target, propKey, receiver);

    if (originalApiClient && typeof originalApiClient === 'object' && propKey in target) {
      return new Proxy(originalApiClient, methodProxyHandler);
    }

    return originalApiClient;
  },
};
