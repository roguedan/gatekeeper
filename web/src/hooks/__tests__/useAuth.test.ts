import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useAuth } from '../useAuth'
import * as authService from '@/services/auth'

vi.mock('@/services/auth', () => ({
  authService: {
    isAuthenticated: vi.fn(),
    getCurrentAddress: vi.fn(),
    logout: vi.fn(),
  },
}))

describe('useAuth', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.spyOn(authService.authService, 'isAuthenticated').mockReturnValue(false)
    vi.spyOn(authService.authService, 'getCurrentAddress').mockReturnValue(null)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('initializes with default state', () => {
    const { result } = renderHook(() => useAuth())

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.address).toBe(null)
    expect(result.current.token).toBe(null)
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBe(null)
  })

  it('sets authenticated state', () => {
    const { result } = renderHook(() => useAuth())

    act(() => {
      result.current.setAuthenticated('0x123', 'token123')
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.address).toBe('0x123')
    expect(result.current.token).toBe('token123')
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBe(null)
  })

  it('sets error state', () => {
    const { result } = renderHook(() => useAuth())

    act(() => {
      result.current.setError('Authentication failed')
    })

    expect(result.current.error).toBe('Authentication failed')
    expect(result.current.isLoading).toBe(false)
  })

  it('sets loading state', () => {
    const { result } = renderHook(() => useAuth())

    act(() => {
      result.current.setLoading(true)
    })

    expect(result.current.isLoading).toBe(true)

    act(() => {
      result.current.setLoading(false)
    })

    expect(result.current.isLoading).toBe(false)
  })

  it('calls logout and resets state', () => {
    const { result } = renderHook(() => useAuth())

    // Set authenticated first
    act(() => {
      result.current.setAuthenticated('0x123', 'token123')
    })

    expect(result.current.isAuthenticated).toBe(true)

    // Logout
    act(() => {
      result.current.logout()
    })

    expect(authService.authService.logout).toHaveBeenCalled()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.address).toBe(null)
    expect(result.current.token).toBe(null)
  })
})
