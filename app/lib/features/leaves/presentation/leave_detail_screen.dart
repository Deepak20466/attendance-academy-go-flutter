import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../auth/data/auth_controller.dart';
import '../data/leave_repository.dart';

class LeaveDetailScreen extends ConsumerStatefulWidget {
  const LeaveDetailScreen({super.key, required this.leaveId});
  final String leaveId;

  @override
  ConsumerState<LeaveDetailScreen> createState() => _LeaveDetailScreenState();
}

class _LeaveDetailScreenState extends ConsumerState<LeaveDetailScreen> {
  bool _busy = false;

  Future<void> _act(Future<void> Function() action) async {
    setState(() => _busy = true);
    try {
      await action();
      if (mounted) Navigator.of(context).maybePop();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Action failed: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    final isAdmin = session?.isAdmin ?? false;
    final repo = ref.watch(leaveRepositoryProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Leave request')),
      body: _busy
          ? const Center(child: CircularProgressIndicator())
          : Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  if (isAdmin) ...[
                    Row(
                      children: [
                        Expanded(
                          child: OutlinedButton(
                            onPressed: () => _act(() => repo.reject(widget.leaveId)),
                            child: const Text('Reject'),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: FilledButton(
                            onPressed: () => _act(() => repo.approve(widget.leaveId)),
                            child: const Text('Approve'),
                          ),
                        ),
                      ],
                    ),
                  ] else
                    OutlinedButton(
                      onPressed: () => _act(() => repo.cancel(widget.leaveId)),
                      child: const Text('Cancel request'),
                    ),
                ],
              ),
            ),
    );
  }
}
