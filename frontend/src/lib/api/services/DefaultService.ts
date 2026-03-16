/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { AuthTokenBody } from '../models/AuthTokenBody';
import type { CreateUrlRequestBody } from '../models/CreateUrlRequestBody';
import type { CreateUrlResponseBody } from '../models/CreateUrlResponseBody';
import type { ErrorModel } from '../models/ErrorModel';
import type { ListUrlsResponseBody } from '../models/ListUrlsResponseBody';
import type { RegisterRequestBody } from '../models/RegisterRequestBody';
import type { UpdateUrlRequestBody } from '../models/UpdateUrlRequestBody';
import type { UpdateUrlResponseBody } from '../models/UpdateUrlResponseBody';
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
     * List authenticated user's shortened URLs
     * @returns ListUrlsResponseBody OK
     * @returns ErrorModel Error
     * @throws ApiError
     */
    public static listUserUrls({
        page = 1,
        pageSize,
    }: {
        page?: number,
        pageSize?: number,
    }): CancelablePromise<ListUrlsResponseBody | ErrorModel> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/user/urls',
            query: {
                'page': page,
                'page_size': pageSize,
            },
        });
    }
    /**
     * Create a new shortened URL
     * @returns ErrorModel Error
     * @returns CreateUrlResponseBody Created
     * @throws ApiError
     */
    public static createShortenedUrl({
        requestBody,
    }: {
        requestBody?: CreateUrlRequestBody,
    }): CancelablePromise<ErrorModel | CreateUrlResponseBody> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/user/urls',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * Delete an existing shortened URL
     * @returns ErrorModel Error
     * @throws ApiError
     */
    public static deleteShortenedUrl({
        id,
        lastUpdated,
    }: {
        id: number,
        lastUpdated: string,
    }): CancelablePromise<ErrorModel> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/user/urls/{id}',
            path: {
                'id': id,
            },
            query: {
                'last_updated': lastUpdated,
            },
        });
    }
    /**
     * Update an existing shortened URL
     * @returns UpdateUrlResponseBody OK
     * @returns ErrorModel Error
     * @throws ApiError
     */
    public static updateShortenedUrl({
        id,
        requestBody,
    }: {
        id: number,
        requestBody?: UpdateUrlRequestBody,
    }): CancelablePromise<UpdateUrlResponseBody | ErrorModel> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/user/urls/{id}',
            path: {
                'id': id,
            },
            body: requestBody,
            mediaType: 'application/json',
        });
    }
}
