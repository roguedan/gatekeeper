import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Alert } from '../Alert'

describe('Alert', () => {
  it('renders children correctly', () => {
    render(<Alert>Alert message</Alert>)
    expect(screen.getByText('Alert message')).toBeInTheDocument()
  })

  it('applies variant styles correctly', () => {
    const { container, rerender } = render(<Alert variant="success">Success</Alert>)
    expect(container.firstChild).toHaveClass('bg-green-50')

    rerender(<Alert variant="error">Error</Alert>)
    expect(container.firstChild).toHaveClass('bg-red-50')

    rerender(<Alert variant="warning">Warning</Alert>)
    expect(container.firstChild).toHaveClass('bg-yellow-50')

    rerender(<Alert variant="info">Info</Alert>)
    expect(container.firstChild).toHaveClass('bg-blue-50')
  })

  it('shows close button when onClose provided', () => {
    const handleClose = vi.fn()
    render(<Alert onClose={handleClose}>Closeable alert</Alert>)

    const closeButton = screen.getByRole('button')
    expect(closeButton).toBeInTheDocument()

    fireEvent.click(closeButton)
    expect(handleClose).toHaveBeenCalledTimes(1)
  })

  it('does not show close button when onClose not provided', () => {
    render(<Alert>Non-closeable alert</Alert>)
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })
})
