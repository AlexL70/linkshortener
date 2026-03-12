/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AuthTokenBody } from '../models/AuthTokenBody';
import type { ErrorModel } from '../models/ErrorModel';
import type { HelloResponseBody } from '../models/HelloResponseBody';
import type { RegisterRequestBody } from '../models/RegisterRequestBody';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DefaultService {
    /**
     * Log out the authenticated user by blacklisting the JWT
     * @returns ErrorModel Error
     * @throws ApiError
     */
    public static logout(): CancelablePromise<ErrorModel> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/auth/logout',
        });
    }
    /**
     * Complete new-user registration after OAuth callback
     * @returns AuthTokenBody OK
     * @returns ErrorModel Error
     * @throws ApiError
     */
    public static registerUser({
        requestBody,
    }: {
        requestBody?: RegisterRequestBody,
    }): CancelablePromise<AuthTokenBody | ErrorModel> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/auth/register',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * Hello World
     * @returns HelloResponseBody OK
     * @returns ErrorModel Error
     * @throws ApiError
     */
    public static hello(): CancelablePromise<HelloResponseBody | ErrorModel> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/hello',
        });
    }
}
