import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/async_value_widget.dart';
import '../../auth/data/auth_controller.dart';
import '../data/leave_repository.dart';

class LeavesScreen extends ConsumerWidget {
  const LeavesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    final isAdmin = session?.isAdmin ?? false;
    final leaves = ref.watch(isAdmin ? allLeavesProvider : myLeavesProvider);

    return Scaffold(
      appBar: AppBar(title: Text(isAdmin ? 'Leave requests' : 'My leaves')),
      floatingActionButton: isAdmin
          ? null
          : FloatingActionButton.extended(
              onPressed: () => _showApplyDialog(context, ref),
              icon: const Icon(Icons.add),
              label: const Text('Apply'),
            ),
      body: RefreshIndicator(
        onRefresh: () => ref.refresh((isAdmin ? allLeavesProvider : myLeavesProvider).future),
        child: AsyncValueWidget(
          value: leaves,
          data: (result) {
            if (result.data.isEmpty) {
              return ListView(
                children: const [
                  Padding(
                    padding: EdgeInsets.all(32),
                    child: Center(child: Text('No leave requests.', style: TextStyle(color: Colors.grey))),
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: result.data.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final leave = result.data[index];
                return ListTile(
                  title: Text('${DateFormat.MMMd().format(leave.startDate)} - ${DateFormat.MMMd().format(leave.endDate)}'),
                  subtitle: Text(leave.reason, maxLines: 1, overflow: TextOverflow.ellipsis),
                  trailing: _statusChip(leave.status),
                  onTap: () async {
                    await context.push('/leaves/${leave.id}');
                    ref.invalidate(isAdmin ? allLeavesProvider : myLeavesProvider);
                  },
                );
              },
            );
          },
        ),
      ),
    );
  }

  Widget _statusChip(String status) {
    final color = switch (status) {
      'approved' => StatusColors.approved,
      'rejected' => StatusColors.rejected,
      'cancelled' => StatusColors.excused,
      _ => StatusColors.pending,
    };
    return Chip(
      label: Text(status[0].toUpperCase() + status.substring(1), style: TextStyle(color: color, fontSize: 12)),
      backgroundColor: color.withValues(alpha: 0.12),
      visualDensity: VisualDensity.compact,
    );
  }

  Future<void> _showApplyDialog(BuildContext context, WidgetRef ref) async {
    DateTime? start;
    DateTime? end;
    final reasonController = TextEditingController();

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Apply for leave'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(start == null ? 'Start date' : DateFormat.yMMMd().format(start!)),
                trailing: const Icon(Icons.calendar_today, size: 18),
                onTap: () async {
                  final picked = await showDatePicker(
                    context: context,
                    initialDate: DateTime.now(),
                    firstDate: DateTime.now(),
                    lastDate: DateTime.now().add(const Duration(days: 365)),
                  );
                  if (picked != null) setState(() => start = picked);
                },
              ),
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(end == null ? 'End date' : DateFormat.yMMMd().format(end!)),
                trailing: const Icon(Icons.calendar_today, size: 18),
                onTap: () async {
                  final picked = await showDatePicker(
                    context: context,
                    initialDate: start ?? DateTime.now(),
                    firstDate: start ?? DateTime.now(),
                    lastDate: DateTime.now().add(const Duration(days: 365)),
                  );
                  if (picked != null) setState(() => end = picked);
                },
              ),
              const SizedBox(height: 8),
              TextField(
                controller: reasonController,
                decoration: const InputDecoration(labelText: 'Reason'),
                maxLines: 2,
              ),
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
            FilledButton(
              onPressed: () async {
                if (start == null || end == null || reasonController.text.trim().isEmpty) return;
                await ref.read(leaveRepositoryProvider).apply(start!, end!, reasonController.text.trim());
                ref.invalidate(myLeavesProvider);
                if (context.mounted) Navigator.pop(context);
              },
              child: const Text('Submit'),
            ),
          ],
        ),
      ),
    );
  }
}
