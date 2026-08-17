export type User = {
  id: string;
  username: string;
  status: 'online_available' | 'online_busy' | 'offline';
};

type UserResponse = { user: User };

export type UserPage = {
  items: User[];
  next_cursor?: string;
};

export type ProblemClass = 'algorithms' | 'concurrency' | 'oop';

export type Invitation = {
  id: string;
  sender: Pick<User, 'id' | 'username'>;
  receiver: Pick<User, 'id' | 'username'>;
  status: 'pending' | 'accepted' | 'declined' | 'expired';
  problem_class: ProblemClass;
  expires_at: string;
};

export type Match = {
  id: string;
  player_one: Pick<User, 'id' | 'username'>;
  player_two: Pick<User, 'id' | 'username'>;
  player_one_score: number;
  player_two_score: number;
  state: 'active' | 'waiting_ready' | 'paused' | 'ended';
  problem_class: ProblemClass;
  problem?: Problem;
  round_number: number;
  round_winner_id?: string;
  player_one_ready: boolean;
  player_two_ready: boolean;
  player_one_skip: boolean;
  player_two_skip: boolean;
  winning_source_code?: string;
  code_snapshots: CodeSnapshot[];
  paused_at?: string;
};

export type CodeSnapshot = {
  user_id: string;
  problem_version_id: string;
  source_code: string;
  revision: number;
  cursor_line: number;
  cursor_column: number;
};

export type Problem = {
  id: string;
  slug: string;
  title: string;
  difficulty: 'easy' | 'medium' | 'hard';
  problem_class: ProblemClass;
  requirements: {
    goroutine?: boolean;
    channel?: boolean;
    wait_group?: boolean;
    mutex?: boolean;
    select?: boolean;
    context_cancel?: boolean;
  };
  statement_markdown: string;
  function_signature: string;
  starter_code: string;
  public_tests: Array<{ arguments?: unknown[]; input?: string; expected: unknown }>;
  time_limit_ms: number;
  memory_limit_mb: number;
};

export type SubmissionStatus =
  | 'queued'
  | 'compiling'
  | 'running'
  | 'accepted'
  | 'wrong_answer'
  | 'compile_error'
  | 'runtime_error'
  | 'time_limit'
  | 'memory_limit'
  | 'internal_error';

export type SubmissionTestCase = {
  kind: 'public' | 'hidden';
  index?: number;
  status: 'passed' | 'failed' | 'not_run';
  input?: string;
  expected?: string;
  actual?: string;
  actual_available?: boolean;
  actual_truncated?: boolean;
  error?: string;
};

export type Submission = {
  id: string;
  match_id: string;
  user_id: string;
  status: SubmissionStatus;
  result?: {
    message?: string;
    duration_ms?: number;
    passed_tests?: number;
    total_tests?: number;
    test_cases?: SubmissionTestCase[];
    console_output?: string;
    console_output_truncated?: boolean;
  };
  created_at: string;
};

export type InvitationState = {
  incoming?: Invitation;
  outgoing?: Invitation;
  match?: Match;
};

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
  }
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...options,
    credentials: 'include',
    headers: {
      ...(options?.body ? { 'Content-Type': 'application/json' } : {}),
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = (await response.json().catch(() => null)) as
      | { code?: string; message?: string }
      | null;
    throw new ApiError(
      response.status,
      error?.code ?? 'REQUEST_FAILED',
      error?.message ?? 'Не удалось выполнить запрос',
    );
  }

  if (response.status === 204) {
    return undefined as T;
  }
  return response.json() as Promise<T>;
}

export async function getMe(): Promise<User | null> {
  try {
    return (await request<UserResponse>('/api/v1/me')).user;
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return null;
    }
    throw error;
  }
}

export async function register(username: string, password: string): Promise<User> {
  return (
    await request<UserResponse>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
  ).user;
}

export async function login(username: string, password: string): Promise<User> {
  return (
    await request<UserResponse>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
  ).user;
}

export async function logout(): Promise<void> {
  await request('/api/v1/auth/logout', { method: 'POST' });
}

export async function heartbeat(): Promise<void> {
  await request('/api/v1/presence/heartbeat', { method: 'POST' });
}

export function listUsers(query: string, cursor: string): Promise<UserPage> {
  const params = new URLSearchParams({ limit: '50' });
  if (query) params.set('q', query);
  if (cursor) params.set('cursor', cursor);
  return request<UserPage>(`/api/v1/users?${params}`);
}

export async function createInvitation(
  receiverId: string,
  problemClass: ProblemClass = 'algorithms',
): Promise<Invitation> {
  return (
    await request<{ invitation: Invitation }>('/api/v1/invitations', {
      method: 'POST',
      body: JSON.stringify({ receiver_id: receiverId, problem_class: problemClass }),
    })
  ).invitation;
}

export function getInvitationState(): Promise<InvitationState> {
  return request<InvitationState>('/api/v1/invitations');
}

export async function acceptInvitation(invitationId: string): Promise<Match> {
  return (
    await request<{ match: Match }>(`/api/v1/invitations/${invitationId}/accept`, {
      method: 'POST',
    })
  ).match;
}

export async function declineInvitation(invitationId: string): Promise<void> {
  await request(`/api/v1/invitations/${invitationId}/decline`, { method: 'POST' });
}

export async function getMatch(matchId: string): Promise<Match> {
  return (await request<{ match: Match }>(`/api/v1/matches/${matchId}`)).match;
}

export async function leaveMatch(matchId: string): Promise<void> {
  await request(`/api/v1/matches/${matchId}/leave`, { method: 'POST' });
}

export async function createSubmission(matchId: string, sourceCode: string): Promise<Submission> {
  return (
    await request<{ submission: Submission }>(`/api/v1/matches/${matchId}/submissions`, {
      method: 'POST',
      body: JSON.stringify({ source_code: sourceCode }),
    })
  ).submission;
}

export async function getSubmission(submissionId: string): Promise<Submission> {
  return (await request<{ submission: Submission }>(`/api/v1/submissions/${submissionId}`))
    .submission;
}

export async function updateCode(
  matchId: string,
  sourceCode: string,
  revision: number,
  cursorLine: number,
  cursorColumn: number,
): Promise<void> {
  await request(`/api/v1/matches/${matchId}/code`, {
    method: 'PUT',
    body: JSON.stringify({
      source_code: sourceCode,
      revision,
      cursor_line: cursorLine,
      cursor_column: cursorColumn,
    }),
  });
}

export async function readyForNextRound(matchId: string): Promise<Match> {
  return (
    await request<{ match: Match }>(`/api/v1/matches/${matchId}/ready`, { method: 'POST' })
  ).match;
}

export async function toggleSkipVote(matchId: string): Promise<Match> {
  return (
    await request<{ match: Match }>(`/api/v1/matches/${matchId}/skip`, { method: 'POST' })
  ).match;
}
