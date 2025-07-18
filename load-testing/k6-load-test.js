import { check, group, sleep } from 'k6';
import http from 'k6/http';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate( 'errors' );
const responseTime = new Trend( 'response_time' );

// Test configuration
export const options = {
    stages: [
        { duration: '2m', target: 10 }, // Ramp up to 10 users
        { duration: '5m', target: 10 }, // Stay at 10 users
        { duration: '2m', target: 20 }, // Ramp up to 20 users
        { duration: '5m', target: 20 }, // Stay at 20 users
        { duration: '2m', target: 0 },  // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
        http_req_failed: ['rate<0.1'],    // Error rate under 10%
        errors: ['rate<0.1'],              // Custom error rate under 10%
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const USERNAME = __ENV.USERNAME || 'john_tech';
const PASSWORD = __ENV.PASSWORD || 'password123';

let authToken = null;

export function setup() {
    // Login and get auth token
    const loginRes = http.post( `${ BASE_URL }/login`, JSON.stringify( {
        username: USERNAME,
        password: PASSWORD,
    } ), {
        headers: { 'Content-Type': 'application/json' },
    } );

    check( loginRes, {
        'login successful': ( r ) => r.status === 200,
        'token received': ( r ) => r.json( 'token' ) !== undefined,
    } );

    return { token: loginRes.json( 'token' ) };
}

export default function ( data ) {
    authToken = data.token;

    const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${ authToken }`,
    };

    group( 'API Load Testing', function () {
        // Test 1: Get all tasks
        group( 'Get Tasks', function () {
            const res = http.get( `${ BASE_URL }/tasks`, { headers } );
            const success = check( res, {
                'get tasks status is 200': ( r ) => r.status === 200,
                'get tasks response time < 500ms': ( r ) => r.timings.duration < 500,
            } );

            errorRate.add( !success );
            responseTime.add( res.timings.duration );
        } );

        // Test 2: Create a task
        group( 'Create Task', function () {
            const taskData = {
                summary: `Load test task ${ Date.now() }`,
                performed_at: new Date().toISOString(),
            };

            const res = http.post( `${ BASE_URL }/tasks`, JSON.stringify( taskData ), { headers } );
            const success = check( res, {
                'create task status is 201': ( r ) => r.status === 201,
                'create task response time < 1000ms': ( r ) => r.timings.duration < 1000,
            } );

            errorRate.add( !success );
            responseTime.add( res.timings.duration );

            if ( success ) {
                const taskId = res.json( 'id' );

                // Test 3: Update the task
                group( 'Update Task', function () {
                    const updateData = {
                        summary: `Updated load test task ${ Date.now() }`,
                        performed_at: new Date().toISOString(),
                    };

                    const updateRes = http.put( `${ BASE_URL }/tasks/${ taskId }`, JSON.stringify( updateData ), { headers } );
                    const updateSuccess = check( updateRes, {
                        'update task status is 200': ( r ) => r.status === 200,
                        'update task response time < 1000ms': ( r ) => r.timings.duration < 1000,
                    } );

                    errorRate.add( !updateSuccess );
                    responseTime.add( updateRes.timings.duration );
                } );
            }
        } );

        // Test 4: Health check endpoint
        group( 'Health Check', function () {
            const res = http.get( `${ BASE_URL }/health/ready` );
            const success = check( res, {
                'health check status is 200': ( r ) => r.status === 200,
                'health check response time < 200ms': ( r ) => r.timings.duration < 200,
            } );

            errorRate.add( !success );
            responseTime.add( res.timings.duration );
        } );
    } );

    sleep( 1 );
}

export function teardown( data ) {
    // Logout or cleanup if needed
    console.log( 'Load test completed' );
}