import { check, group, sleep } from 'k6';
import http from 'k6/http';

// Smoke test configuration - minimal load to verify functionality
export const options = {
    vus: 1, // 1 virtual user
    duration: '30s', // Run for 30 seconds
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
        http_req_failed: ['rate<0.1'],    // Error rate under 10%
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const USERNAME = __ENV.USERNAME || 'john_tech';
const PASSWORD = __ENV.PASSWORD || 'password123';

export default function () {
    // Test 1: Health check
    group( 'Health Check', function () {
        const res = http.get( `${ BASE_URL }/health/ready` );
        check( res, {
            'health check status is 200': ( r ) => r.status === 200,
            'health check response time < 200ms': ( r ) => r.timings.duration < 200,
        } );
    } );

    // Test 2: Login
    group( 'Authentication', function () {
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

        if ( loginRes.status === 200 ) {
            const token = loginRes.json( 'token' );
            const headers = {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${ token }`,
            };

            // Test 3: Get tasks
            group( 'Get Tasks', function () {
                const tasksRes = http.get( `${ BASE_URL }/tasks`, { headers } );
                check( tasksRes, {
                    'get tasks status is 200': ( r ) => r.status === 200,
                    'tasks response is array': ( r ) => Array.isArray( r.json() ),
                } );
            } );

            // Test 4: Create task
            group( 'Create Task', function () {
                const taskData = {
                    summary: 'Smoke test task',
                    performed_at: new Date().toISOString(),
                };

                const createRes = http.post( `${ BASE_URL }/tasks`, JSON.stringify( taskData ), { headers } );
                check( createRes, {
                    'create task status is 201': ( r ) => r.status === 201,
                    'task has id': ( r ) => r.json( 'id' ) !== undefined,
                } );
            } );
        }
    } );

    sleep( 1 );
}