<script lang="ts" setup>
import ArrowTopIcon from '@/components/icons/ArrowTopIcon.vue';
import LinkIcon from '@/components/icons/LinkIcon.vue';
import LoadingIcon from '@/components/icons/LoadingIcon.vue';
import TimesIcon from '@/components/icons/TimesIcon.vue';
import { DashboardTransport } from '@/transport/dashboard';
import { computed, onBeforeMount, ref, watch } from 'vue';

enum TaskStatus {
    Forming = 'Формируется',
    CanDo = 'Можно делать',
    OnHold = 'На паузе',
    Waiting = 'Ожидание',
    InProgress = 'В работе',
    NeedDiscussion = 'Надо обсудить',
    CodeReview = 'Код-ревью',
    InternalCheck = 'Внутренняя проверка',
    ReadyToUpload = 'Можно выгружать',
    ClientCheck = 'Проверка клиентом',
    Cancelled = 'Отменена',
    Done = 'Готова'
}

enum TaskStatusCategory {
    ToDo = 'Необходимо делать',
    InProgress = 'В работе',
    Done = 'Готово'
}

const getCategoryByStatus = (status: TaskStatus): TaskStatusCategory => {
    switch (status) {
        case TaskStatus.Forming:
        case TaskStatus.CanDo:
        case TaskStatus.OnHold:
        case TaskStatus.Waiting:
            return TaskStatusCategory.ToDo;
        case TaskStatus.InProgress:
        case TaskStatus.NeedDiscussion:
        case TaskStatus.CodeReview:
            return TaskStatusCategory.InProgress;
        case TaskStatus.InternalCheck:
        case TaskStatus.ReadyToUpload:
        case TaskStatus.ClientCheck:
        case TaskStatus.Cancelled:
        case TaskStatus.Done:
            return TaskStatusCategory.Done;
        default:
            throw new Error('Unknown task status');
    }
}

interface Task {
    title: string;
    status: TaskStatus;
}

const currentQuarter = Math.ceil((new Date().getMonth() + 1) / 3)

const querterTasks = ref<Task[]>([])
const quarterTasksDone = computed(() => querterTasks.value.filter(task => getCategoryByStatus(task.status) === TaskStatusCategory.Done).length)


onBeforeMount(async () => {
    querterTasks.value = await DashboardTransport.getQuarterTasks()

    const daysInCurrentQuarter = new Date(new Date().getFullYear(), (currentQuarter - 1) * 3, 0).getDate();
    const daysFromStartOfQuarter = new Date().getDate() - new Date(new Date().getFullYear(), (currentQuarter - 1) * 3, 1).getDate();

    const percentOfTasksDone = (querterTasks.value.filter(task => getCategoryByStatus(task.status) === TaskStatusCategory.Done).length / querterTasks.value.length) * 100
    const percentOfDaysGone = (daysFromStartOfQuarter / daysInCurrentQuarter) * 100

    const progressDiff = Math.abs(percentOfTasksDone - percentOfDaysGone)
})

const fileInput = ref<HTMLInputElement | null>(null);
const triggerFileUpload = () => {
    fileInput.value?.click();
};

const handleFileUpload = (event: Event) => {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files.length > 0) {
        const file = target.files[0];
        const reader = new FileReader();
        reader.onload = (e) => {
            mindmapLoaderRequestProducer.start(async () => {
                try {
                    await DashboardTransport.parseMindmap(file)
                    Message.newMessage(MessageType.Success, "Майндкарта загружена")
                } catch (error) {
                    console.error(error)
                    Message.newMessage(MessageType.Error, "Не удалось загрузить майндкарту")

                }
            })
        };
        reader.readAsText(file);
    }
}

enum Period {
    Quarter = 'Квартал',
    Month = 'Месяц',
    Week = 'Неделя'
}

const currentPeriod = ref<Period>(Period.Quarter)

const userTasks = ref<Task[]>([])

const userPercents = ref(0)

watch(currentPeriod, async () => {
    let periodStart: Date;
    let periodEnd: Date;

    const now = new Date();
    switch (currentPeriod.value) {
        case Period.Quarter:
            periodStart = new Date(now.getFullYear(), (currentQuarter - 1) * 3, 1);
            periodEnd = new Date(now.getFullYear(), currentQuarter * 3, 0);
            break;
        case Period.Month:
            periodStart = new Date(now.getFullYear(), now.getMonth(), 1);
            periodEnd = new Date(now.getFullYear(), now.getMonth() + 1, 0);
            break;
        case Period.Week:
            const day = now.getDay();
            const diff = now.getDate() - day + (day === 0 ? -6 : 1);
            periodStart = new Date(now.setDate(diff));
            periodEnd = new Date(now.setDate(diff + 6));
            break;
    }

    progressRequestProducer.start(async () => {
        let tasks;
        try {
            tasks = await DashboardTransport.getTasksOfEmployee('corray9', Math.floor(periodStart.getTime() / 1000), Math.floor(periodEnd.getTime() / 1000))
        } catch (error) {
            console.error(error)
        }
        userTasks.value = tasks


        const now = new Date();
        const totalDaysInPeriod = (periodEnd.getTime() - periodStart.getTime()) / (1000 * 60 * 60 * 24);
        const daysGoneFromPeriodStart = (now.getTime() - periodStart.getTime()) / (1000 * 60 * 60 * 24);

        const percentOfDaysGone = (daysGoneFromPeriodStart / totalDaysInPeriod) * 100;

        const percentOfTasksDone = (userTasks.value.filter(task => getCategoryByStatus(task.status) === TaskStatusCategory.Done).length / userTasks.value.length) * 100 || 0;

        console.log(percentOfDaysGone, percentOfTasksDone);
        userPercents.value = percentOfTasksDone - percentOfDaysGone;

    })

})

class RequestProducer {
    state = ref({
        loading: false
    });

    start = async (f: Function) => {
        this.state.value.loading = true;

        try {
            await f();
        } finally {
            this.state.value.loading = false;
        }
    }
}

const toSheetsRequestProducer = new RequestProducer()

const loadToSheets = async () => {
    try {
        await toSheetsRequestProducer.start(DashboardTransport.updateGoogleSheets);
        Message.newMessage(MessageType.Success, "Данные обновлены")
    } catch (error) {
        console.error(error)
        Message.newMessage(MessageType.Error, "Не удалось обновить данные")
    }
}

const mindmapLoaderRequestProducer = new RequestProducer()

const enum MessageType {
    Error = "error",
    Warning = "warning",
    Success = "success",
    Info = "info"
}

class Message {
    type: MessageType;
    text: string;
    id: number;
    static duration = 3000;

    static messages = ref<Map<number, Message>>(new Map());
    static nextId = 0;

    constructor(type: MessageType, text: string) {
        this.type = type;
        this.text = text;
        this.id = Message.nextId++;
    }

    static newMessage(type: MessageType, text: string) {
        const message = new Message(type, text);
        this.messages.value.set(message.id, message);

        setTimeout(() => {
            this.removeMessage(message.id);
        }, this.duration);
    }

    static removeMessage(id: number) {
        this.messages.value.delete(id);
    }
}

const testMsg = () => {
    Message.newMessage(MessageType.Error, "Упс, произошла ошибка, попробуйте еще раз")
}

const progressRequestProducer = new RequestProducer()

const enum ProgressStatus {
    Good = 'good',
    Bad = 'bad',
    Neutral = 'neutral'
}
const progressToStatus = (num: number) => {
    if (num < 0) return ProgressStatus.Bad
    if (num > 0) return ProgressStatus.Good
    return ProgressStatus.Neutral
}

const salaryNotifyRequestProducer = new RequestProducer()

const salaryNotify = async () => {
    try {
        await salaryNotifyRequestProducer.start(DashboardTransport.notifyAboutSalary);
        Message.newMessage(MessageType.Success, "Уведомления отправлены")
    } catch (error) {
        console.error(error)
        Message.newMessage(MessageType.Error, "Не удалось отправить обновления")
    }
}

</script>

<template>
    <div class="messages">
        <TransitionGroup name="fade-slide">
            <div class="message" v-for="(message, i) of Message.messages.value.values()" :key="`message${i}`"
                @click="Message.removeMessage(message.id)" :class="message.type">
                <div class="msg-icon">
                    <TimesIcon v-if="message.type === MessageType.Error" class=" text-danger" />
                </div>
                <p>{{ message.text }}</p>
            </div>
        </TransitionGroup>
    </div>
    <main class="">
        <div class="header">
            <h1>Incetro</h1>

            <img src="../assets/images/dashboard/incetro-logo.png" alt="">
        </div>

        <div class="dashboard-grid">
            <div @click="loadToSheets" class="card to-sheets">
                <Transition name="fade">
                    <div v-if="toSheetsRequestProducer.state.value.loading" class="card-loader">
                        <div class=" p-2 rounded-xl bg-blue-400 bg-opacity-25">
                            <LoadingIcon class="text-4xl text-primary" />
                        </div>
                        <h4 class="text-center">Бот работает,<br>ожидайте</h4>
                    </div>
                </Transition>
                <div class="card-header">
                    <p>Из Notion в<br>Google Sheets</p>
                    <LinkIcon class="text-2xl" />
                </div>

                <div class="footer text-white">
                    <h3>Выгрузить</h3>
                </div>
            </div>

            <div class="card salary-notify p-default" @click="salaryNotify">
                <Transition name="fade">
                    <div v-if="salaryNotifyRequestProducer.state.value.loading" class="card-loader">
                        <div class=" p-2 rounded-xl bg-blue-400 bg-opacity-25">
                            <LoadingIcon class="text-4xl text-primary" />
                        </div>
                        <h4 class="text-center">Бот работает, ожидайте</h4>
                    </div>
                </Transition>

                <div class="card-header">
                    <p>Провели<br>Выплаты ЗП?</p>
                    <LinkIcon class="text-2xl" />
                </div>

                <div class="footer">
                    <h3>Уведомить</h3>
                </div>
            </div>

            <div class="card qprogress">
                <div class="card-header">
                    <p>Прогресс<br>(Q{{ currentQuarter }})</p>
                    <LinkIcon class="text-2xl" />
                </div>

                <div class="footer">
                    <h3>{{ quarterTasksDone }}/{{ querterTasks.length }}</h3>
                    <div class="progress-bar">
                        <div :style="{ width: `${(quarterTasksDone / querterTasks.length) * 100}%` }"></div>
                    </div>
                </div>

            </div>

            <div class="card progress">
                <Transition name="fade">
                    <div v-if="progressRequestProducer.state.value.loading" class="card-loader">
                        <div class=" p-2 rounded-xl bg-blue-400 bg-opacity-25">
                            <LoadingIcon class="text-4xl text-primary" />
                        </div>
                        <h4 class="text-center">Бот работает, ожидайте</h4>
                    </div>
                </Transition>

                <div class="card-header">
                    <div class=" flex gap-1 items-center">
                        <p>Мой прогресс</p>
                        <p class=" flex items-center" :class="progressToStatus(userPercents)">
                            <ArrowTopIcon class="arrow" /> {{ Math.round(Math.abs(userPercents)) }}%
                        </p>
                    </div>
                </div>

                <div class="period-picker">
                    <div @click="currentPeriod = Period.Quarter" class="period-picker-button"
                        :class="{ active: currentPeriod == Period.Quarter }">
                        Квартал
                    </div>

                    <div @click="currentPeriod = Period.Month" class="period-picker-button"
                        :class="{ active: currentPeriod == Period.Month }">
                        Месяц
                    </div>

                    <div @click="currentPeriod = Period.Week" class="period-picker-button"
                        :class="{ active: currentPeriod == Period.Week }">
                        Неделя
                    </div>
                </div>
                <input type="file" ref="fileInput" @change="handleFileUpload" accept=".md" style="display: none;" />
            </div>

            <div class="card mindmap-loader" @click="triggerFileUpload">
                <Transition name="fade">
                    <div v-if="mindmapLoaderRequestProducer.state.value.loading" class="card-loader">
                        <div class=" p-2 rounded-xl bg-blue-400 bg-opacity-25">
                            <LoadingIcon class="text-4xl text-primary" />
                        </div>
                        <h4 class="text-center">Бот работает,<br>ожидайте</h4>
                    </div>
                </Transition>

                <div class="card-header">
                    <p>Майндкарты</p>
                    <LinkIcon class="text-2xl" />
                </div>

                <div class="footer">
                    <h3>Загрузить</h3>
                </div>
            </div>

            <div class="card feedback" @click="testMsg">
                <div class="card-header">
                    <p>Есть идеи?</p>
                    <LinkIcon class="text-2xl" />
                </div>

                <div class="footer">
                    <h3>Фидбэк</h3>
                </div>
            </div>
        </div>
    </main>
</template>

<style scoped>
.messages {
    @apply fixed bottom-4 w-full h-fit flex flex-col items-center;
}

.message {
    @apply absolute bottom-0 bg-white rounded-2xl flex items-center gap-2 w-full p-4 shadow-lg;
    width: calc(100vw - 2rem);
}

.msg-icon {
    @apply p-2 rounded-xl flex justify-center items-center w-fit h-fit;
}

.message.error>.msg-icon {
    @apply text-danger bg-red-500 bg-opacity-25;
}


.fade-slide-enter-active,
.fade-slide-leave-active {
    transition: all 0.5s ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
    opacity: 0;
    transform: translateY(10px);
}


.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.5s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}

main {
    @apply p-small
}

.header {
    @apply flex justify-between items-center p-default;
}

.header>img {
    @apply h-10 w-10 rounded-full;
}

.dashboard-grid {
    @apply grid grid-cols-2 w-full gap-small;
}

.dashboard-grid>.card {
    @apply rounded-2xl overflow-hidden relative bg-white p-default flex flex-col justify-between gap-8 min-h-40;
}

.dashboard-grid>.card>.card-loader {
    @apply absolute w-full h-full bg-white bg-opacity-75 flex flex-col gap-4 justify-center items-center top-0 left-0;
    backdrop-filter: blur(0.5rem);
}

.card-header {
    @apply flex justify-between;
}

.to-sheets {
    grid-column: 1/2;
    grid-row: 1/3;
    background-image: url(../assets/images/dashboard/bg-waves.png);
    background-size: 125%;
    @apply bg-right-bottom bg-no-repeat;
}

.salary-notify {
    grid-column: 2/3;
    grid-row: 1/2;
}

.qprogress {
    grid-column: 2/3;
    grid-row: 2/3;
}

.progress {
    grid-column: 1/3;
    grid-row: 3/4;
    min-height: auto !important;
}

.progress .good {
    @apply text-success;
}

.progress .bad>.arrow {
    @apply rotate-180;
}

.progress .neutral>.arrow {
    @apply hidden;
}

.progress .bad {
    @apply text-danger;
}

.progress .neutral {
    @apply text-warning;
}

.mindmap-loader {
    grid-column: 1/2;
    grid-row: 4/5;
}

.feedback {
    grid-column: 2/3;
    grid-row: 4/5;
    background-image: url(../assets/images/dashboard/bg-gradient.png);
    @apply bg-cover text-white;
}

.card.progress {
    @apply gap-2;
}

.period-picker {
    @apply flex gap-1 p-1 bg-gray-200 rounded-full;
}

.period-picker>.period-picker-button {
    @apply w-full flex items-center justify-center p-2 rounded-full bg-transparent text-gray-400;
}

.period-picker>.period-picker-button.active {
    @apply bg-white text-black;
}

.progress-bar {
    @apply w-full h-8 bg-gray-300 rounded-full overflow-hidden;
}

.progress-bar>div {
    @apply bg-primary h-full;
}
</style>