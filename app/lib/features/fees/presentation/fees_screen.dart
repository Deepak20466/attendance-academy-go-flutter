import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/async_value_widget.dart';
import '../../activities/data/activity_repository.dart';
import '../../auth/data/auth_controller.dart';
import '../data/fee_repository.dart';

class FeesScreen extends ConsumerStatefulWidget {
  const FeesScreen({super.key});

  @override
  ConsumerState<FeesScreen> createState() => _FeesScreenState();
}

class _FeesScreenState extends ConsumerState<FeesScreen> {
  String? _statusFilter;

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    final isAdmin = session?.isAdmin ?? false;
    final fees = ref.watch(_feesProvider(_statusFilter));

    return Scaffold(
      appBar: AppBar(
        title: const Text('Fees'),
        actions: [
          PopupMenuButton<String?>(
            icon: const Icon(Icons.filter_list),
            onSelected: (v) => setState(() => _statusFilter = v),
            itemBuilder: (context) => const [
              PopupMenuItem(value: null, child: Text('All')),
              PopupMenuItem(value: 'pending', child: Text('Pending')),
              PopupMenuItem(value: 'paid', child: Text('Paid')),
              PopupMenuItem(value: 'overdue', child: Text('Overdue')),
            ],
          ),
        ],
      ),
      floatingActionButton: isAdmin
          ? FloatingActionButton.extended(
              onPressed: () => _showGenerateDialog(context, ref),
              icon: const Icon(Icons.add),
              label: const Text('Generate fees'),
            )
          : null,
      body: RefreshIndicator(
        onRefresh: () => ref.refresh(_feesProvider(_statusFilter).future),
        child: AsyncValueWidget(
          value: fees,
          data: (result) {
            if (result.data.isEmpty) {
              return ListView(
                children: const [
                  Padding(
                    padding: EdgeInsets.all(32),
                    child: Center(child: Text('No fee records found.', style: TextStyle(color: Colors.grey))),
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: result.data.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final fee = result.data[index];
                final color = switch (fee.status) {
                  'paid' => StatusColors.approved,
                  'overdue' => StatusColors.rejected,
                  _ => StatusColors.pending,
                };
                return ListTile(
                  title: Text('${fee.amount.toStringAsFixed(2)} • ${DateFormat.MMMM().format(DateTime(0, fee.periodMonth))} ${fee.periodYear}'),
                  subtitle: Text('Due ${DateFormat.yMMMd().format(fee.dueDate)}'),
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Chip(
                        label: Text(fee.status, style: TextStyle(color: color, fontSize: 12)),
                        backgroundColor: color.withValues(alpha: 0.12),
                        visualDensity: VisualDensity.compact,
                      ),
                      if (fee.status != 'paid' && isAdmin)
                        TextButton(
                          onPressed: () async {
                            await ref.read(feeRepositoryProvider).markPaid(fee.id);
                            ref.invalidate(_feesProvider(_statusFilter));
                          },
                          child: const Text('Mark paid'),
                        ),
                    ],
                  ),
                );
              },
            );
          },
        ),
      ),
    );
  }

  Future<void> _showGenerateDialog(BuildContext context, WidgetRef ref) async {
    final activities = await ref.read(activityRepositoryProvider).list();
    if (!context.mounted) return;
    if (activities.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Create an activity first.')),
      );
      return;
    }

    final now = DateTime.now();
    String selectedActivityId = activities.first.id;
    final amountController = TextEditingController();
    DateTime dueDate = DateTime(now.year, now.month, 10);
    String? errorText;

    await showDialog<void>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Generate monthly fees'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text('Creates a pending fee for every active student in the activity for ${DateFormat.MMMM().format(now)} ${now.year}.',
                  style: const TextStyle(color: Colors.grey, fontSize: 13)),
              const SizedBox(height: 16),
              DropdownButtonFormField<String>(
                initialValue: selectedActivityId,
                decoration: const InputDecoration(labelText: 'Activity'),
                items: activities.map((a) => DropdownMenuItem(value: a.id, child: Text(a.name))).toList(),
                onChanged: (v) => setState(() => selectedActivityId = v!),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: amountController,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(labelText: 'Amount per student'),
              ),
              const SizedBox(height: 12),
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text('Due date: ${DateFormat.yMMMd().format(dueDate)}'),
                trailing: const Icon(Icons.calendar_today, size: 18),
                onTap: () async {
                  final picked = await showDatePicker(
                    context: context,
                    initialDate: dueDate,
                    firstDate: DateTime(now.year - 1),
                    lastDate: DateTime(now.year + 1),
                  );
                  if (picked != null) setState(() => dueDate = picked);
                },
              ),
              if (errorText != null) ...[
                const SizedBox(height: 8),
                Text(errorText!, style: const TextStyle(color: Colors.red)),
              ],
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
            FilledButton(
              onPressed: () async {
                final amount = double.tryParse(amountController.text);
                if (amount == null || amount <= 0) {
                  setState(() => errorText = 'Enter a valid amount.');
                  return;
                }
                try {
                  final count = await ref.read(feeRepositoryProvider).generate(
                        activityId: selectedActivityId,
                        amount: amount,
                        dueDate: dueDate,
                        month: now.month,
                        year: now.year,
                      );
                  ref.invalidate(_feesProvider(_statusFilter));
                  if (context.mounted) {
                    Navigator.pop(context);
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Created $count fee record(s).')),
                    );
                  }
                } catch (e) {
                  setState(() => errorText = 'Failed: $e');
                }
              },
              child: const Text('Generate'),
            ),
          ],
        ),
      ),
    );
  }
}

final _feesProvider = FutureProvider.autoDispose.family((ref, String? status) {
  return ref.watch(feeRepositoryProvider).list(status: status);
});
