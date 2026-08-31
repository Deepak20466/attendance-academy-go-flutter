import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../auth/data/auth_controller.dart';
import '../../coaches/data/coach_repository.dart';
import '../data/attendance_repository.dart';

class CoachAttendanceScreen extends ConsumerStatefulWidget {
  const CoachAttendanceScreen({super.key});

  @override
  ConsumerState<CoachAttendanceScreen> createState() => _CoachAttendanceScreenState();
}

class _CoachAttendanceScreenState extends ConsumerState<CoachAttendanceScreen> {
  String? _coachFilter;

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(authControllerProvider).valueOrNull;
    final isAdmin = session?.isAdmin ?? false;
    final history = ref.watch(_historyProvider(_coachFilter));

    return Scaffold(
      appBar: AppBar(title: const Text('Coach Attendance')),
      body: Column(
        children: [
          if (isAdmin)
            Padding(
              padding: const EdgeInsets.all(16),
              child: _CoachFilterDropdown(
                selected: _coachFilter,
                onChanged: (v) => setState(() => _coachFilter = v),
              ),
            ),
          Expanded(
            child: RefreshIndicator(
              onRefresh: () => ref.refresh(_historyProvider(_coachFilter).future),
              child: AsyncValueWidget(
                value: history,
                data: (result) {
                  if (result.data.isEmpty) {
                    return ListView(
                      children: const [
                        Padding(
                          padding: EdgeInsets.all(32),
                          child: Center(child: Text('No check-in records found.', style: TextStyle(color: Colors.grey))),
                        ),
                      ],
                    );
                  }
                  return ListView.separated(
                    itemCount: result.data.length,
                    separatorBuilder: (_, __) => const Divider(height: 1),
                    itemBuilder: (context, index) {
                      final r = result.data[index];
                      return ListTile(
                        leading: Icon(
                          r.checkOutTime != null ? Icons.check_circle_outline : Icons.schedule_outlined,
                          color: r.checkOutTime != null ? Colors.green : Colors.orange,
                        ),
                        title: Text(DateFormat.yMMMd().format(r.attendanceDate)),
                        subtitle: Text(_subtitle(r)),
                        trailing: !r.checkInVerified || (r.checkOutTime != null && !r.checkOutVerified)
                            ? const Tooltip(
                                message: 'Location could not be verified',
                                child: Icon(Icons.location_off_outlined, color: Colors.red, size: 20),
                              )
                            : const Icon(Icons.location_on_outlined, color: Colors.green, size: 20),
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

  String _subtitle(CoachAttendanceRecord r) {
    final inTime = r.checkInTime != null ? DateFormat.jm().format(r.checkInTime!) : '—';
    final outTime = r.checkOutTime != null ? DateFormat.jm().format(r.checkOutTime!) : 'not checked out';
    return 'In: $inTime  •  Out: $outTime';
  }
}

class _CoachFilterDropdown extends ConsumerWidget {
  const _CoachFilterDropdown({required this.selected, required this.onChanged});
  final String? selected;
  final ValueChanged<String?> onChanged;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final coaches = ref.watch(_coachesListProvider);
    return coaches.when(
      loading: () => const LinearProgressIndicator(),
      error: (e, __) => Text('Failed to load coaches: $e'),
      data: (result) => DropdownButtonFormField<String?>(
        initialValue: selected,
        decoration: const InputDecoration(labelText: 'Coach'),
        items: [
          const DropdownMenuItem<String?>(value: null, child: Text('All coaches')),
          ...result.data.map((c) => DropdownMenuItem<String?>(value: c.id, child: Text(c.name))),
        ],
        onChanged: onChanged,
      ),
    );
  }
}

final _coachesListProvider = FutureProvider.autoDispose((ref) {
  return ref.watch(coachRepositoryProvider).list(pageSize: 100);
});

final _historyProvider = FutureProvider.autoDispose.family((ref, String? coachId) {
  return ref.watch(attendanceRepositoryProvider).coachHistory(coachId: coachId, pageSize: 50);
});
