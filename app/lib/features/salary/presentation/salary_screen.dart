import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/api/paged_result.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/async_value_widget.dart';
import '../../auth/data/auth_controller.dart';
import '../data/salary_models.dart';
import '../data/salary_repository.dart';

typedef _Period = ({int month, int year});

class SalaryScreen extends ConsumerStatefulWidget {
  const SalaryScreen({super.key});

  @override
  ConsumerState<SalaryScreen> createState() => _SalaryScreenState();
}

class _SalaryScreenState extends ConsumerState<SalaryScreen> {
  DateTime _period = DateTime(DateTime.now().year, DateTime.now().month);

  void _shiftMonth(int delta) {
    setState(() => _period = DateTime(_period.year, _period.month + delta));
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    final isAdmin = session?.isAdmin ?? false;
    final period = (month: _period.month, year: _period.year);
    final acks = ref.watch(isAdmin ? _periodProvider(period) : _mineProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Salary')),
      floatingActionButton: isAdmin
          ? FloatingActionButton.extended(
              onPressed: () async {
                final count = await ref.read(salaryRepositoryProvider).generate(period.month, period.year);
                ref.invalidate(_periodProvider(period));
                if (context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Created $count acknowledgement(s) for ${DateFormat.yMMMM().format(_period)}.')),
                  );
                }
              },
              icon: const Icon(Icons.add),
              label: Text('Generate ${DateFormat.MMMM().format(_period)}'),
            )
          : null,
      body: Column(
        children: [
          if (isAdmin)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  IconButton(onPressed: () => _shiftMonth(-1), icon: const Icon(Icons.chevron_left)),
                  SizedBox(
                    width: 160,
                    child: Text(DateFormat.yMMMM().format(_period), textAlign: TextAlign.center, style: const TextStyle(fontWeight: FontWeight.w600)),
                  ),
                  IconButton(onPressed: () => _shiftMonth(1), icon: const Icon(Icons.chevron_right)),
                ],
              ),
            ),
          Expanded(
            child: RefreshIndicator(
              onRefresh: () => ref.refresh((isAdmin ? _periodProvider(period) : _mineProvider).future),
              child: AsyncValueWidget(
                value: acks,
                data: (result) {
                  if (result.data.isEmpty) {
                    return ListView(
                      children: const [
                        Padding(
                          padding: EdgeInsets.all(32),
                          child: Center(child: Text('No salary records found.', style: TextStyle(color: Colors.grey))),
                        ),
                      ],
                    );
                  }
                  return ListView.separated(
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    itemCount: result.data.length,
                    separatorBuilder: (_, __) => const Divider(height: 1),
                    itemBuilder: (context, index) {
                      final ack = result.data[index];
                      final acknowledged = ack.status == 'acknowledged';
                      return ListTile(
                        title: Text('${DateFormat.MMMM().format(DateTime(0, ack.periodMonth))} ${ack.periodYear}'),
                        subtitle: Text(ack.amount != null ? 'Amount: ${ack.amount!.toStringAsFixed(2)}' : ''),
                        trailing: acknowledged || isAdmin
                            ? Chip(
                                label: Text(
                                  ack.status,
                                  style: TextStyle(color: acknowledged ? StatusColors.approved : StatusColors.pending, fontSize: 12),
                                ),
                                backgroundColor: (acknowledged ? StatusColors.approved : StatusColors.pending).withValues(alpha: 0.12),
                                visualDensity: VisualDensity.compact,
                              )
                            : FilledButton(
                                onPressed: () async {
                                  await ref.read(salaryRepositoryProvider).acknowledge(ack.id);
                                  ref.invalidate(_mineProvider);
                                },
                                child: const Text('Acknowledge'),
                              ),
                      );
                    },
                  );
                },
              ),
            ),
          ),
        ],
      ),
    );
  }
}

final _mineProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(salaryRepositoryProvider).mine();
});

final _periodProvider = FutureProvider.autoDispose.family<PagedResult<SalaryAckModel>, _Period>((ref, period) {
  return ref.watch(salaryRepositoryProvider).forPeriod(period.month, period.year);
});
