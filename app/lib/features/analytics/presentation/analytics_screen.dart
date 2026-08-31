import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/widgets/async_value_widget.dart';
import '../../../core/widgets/stat_card.dart';
import '../../activities/data/activity_repository.dart';
import '../data/analytics_models.dart';
import '../data/analytics_repository.dart';

typedef _ActivityPeriod = ({String activityId, int month, int year});

class AnalyticsScreen extends ConsumerStatefulWidget {
  const AnalyticsScreen({super.key});

  @override
  ConsumerState<AnalyticsScreen> createState() => _AnalyticsScreenState();
}

class _AnalyticsScreenState extends ConsumerState<AnalyticsScreen> {
  String? _selectedActivityId;
  DateTime _period = DateTime(DateTime.now().year, DateTime.now().month);

  void _shiftMonth(int delta) {
    setState(() => _period = DateTime(_period.year, _period.month + delta));
  }

  @override
  Widget build(BuildContext context) {
    final activities = ref.watch(activitiesListProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Analytics')),
      body: AsyncValueWidget(
        value: activities,
        data: (activityList) {
          if (activityList.isEmpty) {
            return const Center(child: Text('No activities yet.', style: TextStyle(color: Colors.grey)));
          }
          _selectedActivityId ??= activityList.first.id;

          final period = (activityId: _selectedActivityId!, month: _period.month, year: _period.year);
          final summary = ref.watch(_activitySummaryProvider(period));

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              DropdownButtonFormField<String>(
                initialValue: _selectedActivityId,
                decoration: const InputDecoration(labelText: 'Activity'),
                items: activityList.map((a) => DropdownMenuItem(value: a.id, child: Text(a.name))).toList(),
                onChanged: (v) => setState(() => _selectedActivityId = v),
              ),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  IconButton(onPressed: () => _shiftMonth(-1), icon: const Icon(Icons.chevron_left)),
                  SizedBox(
                    width: 160,
                    child: Text(
                      DateFormat.yMMMM().format(_period),
                      textAlign: TextAlign.center,
                      style: const TextStyle(fontWeight: FontWeight.w600),
                    ),
                  ),
                  IconButton(onPressed: () => _shiftMonth(1), icon: const Icon(Icons.chevron_right)),
                ],
              ),
              const SizedBox(height: 8),
              AsyncValueWidget(
                value: summary,
                data: (s) => Column(
                  children: [
                    GridView.count(
                      crossAxisCount: 2,
                      shrinkWrap: true,
                      physics: const NeverScrollableScrollPhysics(),
                      mainAxisSpacing: 12,
                      crossAxisSpacing: 12,
                      childAspectRatio: 1.4,
                      children: [
                        StatCard(label: 'Students', value: '${s.studentCount}', icon: Icons.groups_outlined),
                        StatCard(label: 'Coaches', value: '${s.coachCount}', icon: Icons.badge_outlined),
                        StatCard(label: 'Classes this month', value: '${s.classCount}', icon: Icons.event_note_outlined),
                        StatCard(
                          label: '100% attendance',
                          value: '${s.perfectAttendance}',
                          icon: Icons.workspace_premium_outlined,
                          color: Colors.amber.shade700,
                        ),
                        StatCard(
                          label: 'Fees collected',
                          value: s.feesCollected.toStringAsFixed(0),
                          icon: Icons.payments_outlined,
                          color: Colors.green,
                        ),
                        StatCard(
                          label: 'Fees pending',
                          value: s.feesPending.toStringAsFixed(0),
                          icon: Icons.pending_actions_outlined,
                          color: Colors.red,
                        ),
                        StatCard(
                          label: 'Coach attendance (days)',
                          value: '${s.coachAttendanceDays}',
                          icon: Icons.location_on_outlined,
                          color: Colors.teal,
                        ),
                      ],
                    ),
                    const SizedBox(height: 24),
                    Text('Attendance breakdown', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700)),
                    const SizedBox(height: 12),
                    SizedBox(
                      height: 180,
                      child: (s.presentCount + s.absentCount) == 0
                          ? const Center(child: Text('No attendance data yet.', style: TextStyle(color: Colors.grey)))
                          : PieChart(
                              PieChartData(
                                sectionsSpace: 2,
                                centerSpaceRadius: 40,
                                sections: [
                                  PieChartSectionData(
                                    value: s.presentCount.toDouble(),
                                    color: Colors.green,
                                    title: 'Present\n${s.presentCount}',
                                    radius: 60,
                                    titleStyle: const TextStyle(fontSize: 11, color: Colors.white, fontWeight: FontWeight.w600),
                                  ),
                                  PieChartSectionData(
                                    value: s.absentCount.toDouble(),
                                    color: Colors.red,
                                    title: 'Absent\n${s.absentCount}',
                                    radius: 60,
                                    titleStyle: const TextStyle(fontSize: 11, color: Colors.white, fontWeight: FontWeight.w600),
                                  ),
                                ],
                              ),
                            ),
                    ),
                    const SizedBox(height: 24),
                    Align(
                      alignment: Alignment.centerLeft,
                      child: Text('Students with 100% attendance', style: Theme.of(context).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700)),
                    ),
                    const SizedBox(height: 12),
                    _PerfectAttendanceList(period: period),
                  ],
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _PerfectAttendanceList extends ConsumerWidget {
  const _PerfectAttendanceList({required this.period});
  final _ActivityPeriod period;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final list = ref.watch(_perfectAttendanceProvider(period));
    return AsyncValueWidget(
      value: list,
      data: (students) {
        if (students.isEmpty) {
          return const Text('No students with perfect attendance this month.', style: TextStyle(color: Colors.grey));
        }
        return Column(
          children: students
              .map((s) => ListTile(
                    contentPadding: EdgeInsets.zero,
                    leading: const Icon(Icons.star, color: Colors.amber),
                    title: Text(s.studentName),
                    trailing: Text('${s.totalClasses} classes'),
                  ))
              .toList(),
        );
      },
    );
  }
}

final _activitySummaryProvider = FutureProvider.autoDispose.family<ActivitySummary, _ActivityPeriod>((ref, period) {
  return ref.watch(analyticsRepositoryProvider).activitySummary(period.activityId, month: period.month, year: period.year);
});

final _perfectAttendanceProvider = FutureProvider.autoDispose.family<List<StudentAttendanceSummary>, _ActivityPeriod>((ref, period) {
  return ref.watch(analyticsRepositoryProvider).perfectAttendance(period.activityId, month: period.month, year: period.year);
});
